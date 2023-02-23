package adapter

import (
	"encoding/base64"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/dysodeng/filesystem/storage"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// AliOssAdapter 阿里云OSS存储适配器
type AliOssAdapter struct {
	client *oss.Client
	bucket *oss.Bucket
	config AliOssConfig
}

type AliOssConfig struct {
	AccessId       string
	AccessKey      string
	EndPoint       string
	Region         string
	BucketName     string
	StayBucketName string
	StsRoleArn     string
	IsPrivate      bool
}

func NewAliOssAdapter(config AliOssConfig) Adapter {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	aliOssAdapter := new(AliOssAdapter)

	client, err := oss.New(config.EndPoint, config.AccessId, config.AccessKey)
	if err != nil {
		panic("ali_oss connect error:" + err.Error())
	}

	aliOssAdapter.client = client

	bucket, err := aliOssAdapter.client.Bucket(config.BucketName)
	if err != nil {
		panic("ali_oss bucket error:" + err.Error())
	}

	aliOssAdapter.bucket = bucket
	aliOssAdapter.config = config

	return aliOssAdapter
}

// Info 文件信息
func (adapter *AliOssAdapter) Info(file string) (storage.Attribute, error) {
	res, err := adapter.bucket.GetObjectDetailedMeta(file)
	if err != nil {
		return nil, FileNotExists
	}

	lastModified, _ := time.Parse(time.RFC1123, res["Last-Modified"][0])
	fileSize, _ := strconv.ParseInt(res["Content-Length"][0], 10, 64)

	return storage.NewFileAttribute(file, "", res["Content-Type"][0], fileSize, lastModified.In(time.Local).Unix()), nil
}

func (adapter *AliOssAdapter) HasFile(file string) bool {
	result, err := adapter.bucket.IsObjectExist(file)
	if err != nil {
		return false
	}
	return result
}

func (adapter *AliOssAdapter) HasDir(file string) bool {
	return true
}

func (adapter *AliOssAdapter) Read(file string) (io.ReadCloser, error) {
	return adapter.bucket.GetObject(file)
}

func (adapter *AliOssAdapter) Save(dstFile string, srcFile io.Reader, mimeType string) (bool, error) {
	var options []oss.Option
	if mimeType != "" {
		options = []oss.Option{
			oss.ContentType(mimeType),
		}
	}

	if err := adapter.bucket.PutObject(dstFile, srcFile, options...); err != nil {
		return false, err
	}

	return true, nil
}

func (adapter *AliOssAdapter) Copy(srcFile, disFile string) (bool, error) {
	options := []oss.Option{
		// 复制元数据
		oss.MetadataDirective(oss.MetaCopy),
		// 禁止覆盖目标同名文件
		oss.ForbidOverWrite(true),
		// 标准存储
		oss.StorageClass("Standard"),
	}

	_, err := adapter.bucket.CopyObject(srcFile, disFile, options...)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (adapter *AliOssAdapter) Move(srcFile, disFile string) (bool, error) {
	if _, err := adapter.Copy(srcFile, disFile); err != nil {
		return false, err
	}
	if _, err := adapter.Delete(srcFile); err != nil {
		return false, err
	}
	return true, nil
}

func (adapter *AliOssAdapter) Cover(sourceImagePath, coverImagePath string, width, height uint) error {
	style := "image/resize,m_lfit"
	if width > 0 {
		style += ",w_" + strconv.Itoa(int(width))
	}
	if height > 0 {
		style += ",h_" + strconv.Itoa(int(height))
	}
	process := fmt.Sprintf("%s|sys/saveas,o_%v", style, base64.URLEncoding.EncodeToString([]byte(coverImagePath)))

	_, err := adapter.bucket.ProcessObject(sourceImagePath, process)
	if err != nil {
		log.Println(err)
		return errors.New("缩略图生成失败")
	}

	return nil
}

func (adapter *AliOssAdapter) Delete(file string) (bool, error) {
	if err := adapter.bucket.DeleteObject(file); err != nil {
		return false, err
	}
	return true, nil
}

func (adapter *AliOssAdapter) MultipleDelete(fileList []string) (bool, error) {
	_, err := adapter.bucket.DeleteObjects(fileList)
	oss.DeleteObjectsQuiet(true)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (adapter *AliOssAdapter) MkDir(dir string, mode os.FileMode) (bool, error) {
	return true, nil
}

func (adapter *AliOssAdapter) DeleteDir(dir string) (bool, error) {
	return true, nil
}

func (adapter *AliOssAdapter) List(dir string, iterable func(attribute storage.Attribute)) error {
	dir = strings.TrimRight(dir, "/") + "/"
	prefixDir := oss.Prefix(dir)
	continueToken := ""

	lsRes, err := adapter.bucket.ListObjectsV2(prefixDir, oss.ContinuationToken(continueToken), oss.Delimiter("/"), oss.FetchOwner(true))
	if err != nil {
		return err
	}

	for _, prefix := range lsRes.CommonPrefixes {
		iterable(storage.NewDirectoryAttribute(prefix, "", 0))
	}
	for _, object := range lsRes.Objects {
		if object.Key == dir {
			continue
		}
		iterable(storage.NewFileAttribute(object.Key, "", "", object.Size, object.LastModified.Unix()))
	}

	return nil
}

func (adapter *AliOssAdapter) FullPath(path string) string {
	if adapter.config.IsPrivate {
		signUrl, err := adapter.bucket.SignURL(path, oss.HTTPGet, 60+8*3600)
		if err != nil {
			return ""
		}
		return strings.Replace(signUrl, "http://", "https://", 1)
	}
	return "https://" + adapter.config.BucketName + "." + adapter.config.EndPoint + "/" + path
}

func (adapter *AliOssAdapter) OriginalPath(fullPath string) string {
	u, err := url.Parse(fullPath)
	if err != nil {
		return fullPath
	}
	return strings.TrimLeft(u.Path, "/")
}
