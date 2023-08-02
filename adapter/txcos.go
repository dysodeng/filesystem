package adapter

import (
	"context"
	"fmt"
	"github.com/dysodeng/filesystem/storage"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// TxCosAdapter 腾讯云COS存储适配器
type TxCosAdapter struct {
	client *cos.Client
	config TxCosConfig
}

type TxCosConfig struct {
	SecretID   string
	SecretKey  string
	Token      string
	Region     string
	BucketName string
	IsPrivate  bool
}

func NewTxCosAdapter(config TxCosConfig) Adapter {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.BucketName, config.Region))
	if err != nil {
		panic("tx cos connect error:" + err.Error())
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:     config.SecretID,
			SecretKey:    config.SecretKey,
			SessionToken: config.Token,
		},
	})

	return &TxCosAdapter{
		client: client,
		config: config,
	}
}

// Info 文件/目录信息
// @param file string 文件路径
func (adapter *TxCosAdapter) Info(file string) (storage.Attribute, error) {
	res, err := adapter.client.Object.Head(context.Background(), file, nil)
	if err != nil {
		return nil, FileNotExists
	}

	lastModified, _ := time.Parse(time.RFC1123, res.Header.Get("Last-Modified"))
	fileSize, _ := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
	contentType := res.Header.Get("Content-Type")

	names := strings.Split(strings.TrimRight(file, "/"), "/")

	return storage.NewFileAttribute(names[len(names)-1], file, "", contentType, fileSize, lastModified.In(time.Local).Unix()), nil
}

// HasFile 判断文件是否存在
// @param file string 文件路径
func (adapter *TxCosAdapter) HasFile(file string) bool {
	res, err := adapter.client.Object.IsExist(context.Background(), file)
	if err != nil {
		return false
	}
	return res
}

// HasDir 判断目录是否存在
// @param file string 文件路径
func (adapter *TxCosAdapter) HasDir(file string) bool {
	return true
}

// Read 读取文件内容
// @param file string 文件路径
func (adapter *TxCosAdapter) Read(file string) (io.ReadCloser, error) {
	res, err := adapter.client.Object.Get(context.Background(), file, nil)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

// Save 保存文件
// @param dstFile string 目标文件路径
// @param srcFile io.Reader 原文件内容
func (adapter *TxCosAdapter) Save(dstFile string, srcFile io.Reader, mimeType string) (bool, error) {
	opt := &cos.ObjectPutOptions{}
	if mimeType != "" {
		opt.ObjectPutHeaderOptions = &cos.ObjectPutHeaderOptions{
			ContentType: mimeType,
		}
	}

	_, err := adapter.client.Object.Put(context.Background(), dstFile, srcFile, opt)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Cover 生成缩略图封面
// @param sourceImagePath string 原文件路径
// @param coverImagePath string 目标文件路径
func (adapter *TxCosAdapter) Cover(sourceImagePath, coverImagePath string, width, height uint) error {
	operation := "imageMogr2/thumbnail/"
	if width > 0 {
		operation += fmt.Sprintf("%dx", width)
	}
	if height > 0 {
		if width > 0 {
			operation += fmt.Sprintf("%d", height)
		} else {
			operation += fmt.Sprintf("x%d", height)
		}
	}

	res, err := adapter.client.CI.Get(context.Background(), sourceImagePath, operation, nil)
	if err != nil {
		return FileNotExists
	}

	_, err = adapter.Save(coverImagePath, res.Body, res.Header.Get("Content-Type"))

	return err
}

// Copy 复制文件/目录
// @param srcFile string 原文件路径
// @param dstFile string 目标文件路径
func (adapter *TxCosAdapter) Copy(srcFile, disFile string) (bool, error) {
	srcFileUrl := fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", adapter.config.BucketName, adapter.config.Region, srcFile)
	_, _, err := adapter.client.Object.Copy(context.Background(), disFile, srcFileUrl, nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Move 移动文件/目录
// @param dstFile string 目标文件路径
// @param srcFile string 原文件路径
func (adapter *TxCosAdapter) Move(dstFile, srcFile string) (bool, error) {
	_, err := adapter.Copy(srcFile, dstFile)
	if err != nil {
		return false, err
	}
	return adapter.Delete(srcFile)
}

// Delete 删除文件
// @param file string 文件路径
func (adapter *TxCosAdapter) Delete(file string) (bool, error) {
	_, err := adapter.client.Object.Delete(context.Background(), file)
	if err != nil {
		return false, err
	}
	return true, nil
}

// MultipleDelete 删除多个文件
// @param fileList []string 文件列表
func (adapter *TxCosAdapter) MultipleDelete(fileList []string) (bool, error) {
	var objects []cos.Object
	for _, s := range fileList {
		objects = append(objects, cos.Object{Key: s})
	}

	_, _, err := adapter.client.Object.DeleteMulti(context.Background(), &cos.ObjectDeleteMultiOptions{
		Objects: objects,
		Quiet:   true,
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

// MkDir 创建目录
// @param dir string 目录路径
func (adapter *TxCosAdapter) MkDir(dir string, mode os.FileMode) (bool, error) {
	return true, nil
}

// DeleteDir 删除目录
// @param dir string 目录路径
func (adapter *TxCosAdapter) DeleteDir(dir string) (bool, error) {
	return true, nil
}

// List 文件/目录列表
// @param dir string 目录路径
// @param iterable func 迭代器
func (adapter *TxCosAdapter) List(dir string, iterable func(attribute storage.Attribute)) error {
	var prefixDir string
	if dir == "" || dir == "/" {
		prefixDir = ""
	} else {
		prefixDir = strings.TrimRight(dir, "/") + "/"
	}

	var marker string
	opt := &cos.BucketGetOptions{
		Prefix:    prefixDir,
		Delimiter: "/",
		MaxKeys:   1000,
	}

	isTruncated := true

	for isTruncated {
		opt.Marker = marker
		v, _, err := adapter.client.Bucket.Get(context.Background(), opt)
		if err != nil {
			fmt.Println(err)
			break
		}

		for _, commonPrefix := range v.CommonPrefixes {
			names := strings.Split(strings.TrimRight(commonPrefix, "/"), "/")
			iterable(storage.NewDirectoryAttribute(names[len(names)-1], commonPrefix, "", 0))
		}

		for _, content := range v.Contents {
			if content.Key == dir {
				continue
			}
			lastModified, _ := time.Parse("2006-01-02T15:04:05.000Z", content.LastModified)
			names := strings.Split(strings.TrimRight(content.Key, "/"), "/")
			iterable(storage.NewFileAttribute(names[len(names)-1], content.Key, "", "", content.Size, lastModified.Local().Unix()))
		}

		isTruncated = v.IsTruncated
		marker = v.NextMarker

	}

	return nil
}

// FullPath 获取全路径
// @param path string 文件路径
func (adapter *TxCosAdapter) FullPath(path string) string {
	if adapter.config.IsPrivate {
		return adapter.client.Object.GetObjectURL(path).String()
	}
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", adapter.config.BucketName, adapter.config.Region, path)
}

// OriginalPath 获取原始路径
// @param fullPath string 文件全路径
func (adapter *TxCosAdapter) OriginalPath(fullPath string) string {
	u, err := url.Parse(fullPath)
	if err != nil {
		return fullPath
	}
	return strings.TrimLeft(u.Path, "/")
}
