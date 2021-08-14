package storage

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pkg/errors"
)

// AliOssStorage 阿里云存储
type AliOssStorage struct {
	client     *oss.Client
	bucket     *oss.Bucket
	endpoint   string
	bucketName string
}

type AliOssConfig struct {
	AccessId         string
	AccessKey        string
	EndPoint         string
	Region           string
	BucketName       string
	StayBucketName	 string
	StsRoleArn       string
}

// NewAliOssStorage create ali_oss storage
func NewAliOssStorage(config AliOssConfig) Storage {

	aliStorage := new(AliOssStorage)

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	client, err := oss.New(config.EndPoint, config.AccessId, config.AccessKey)
	if err != nil {
		panic("ali_oss connect error:" + err.Error())
	}

	aliStorage.client = client

	bucket, err := client.Bucket(config.BucketName)
	if err != nil {
		panic("ali_oss bucket error:" + err.Error())
	}
	aliStorage.bucket = bucket

	aliStorage.bucketName = config.BucketName
	aliStorage.endpoint = config.EndPoint

	var storage Storage = aliStorage

	return storage
}

// HasFile 判断文件是否存在
func (storage *AliOssStorage) HasFile(filePath string) bool {

	result, err := storage.bucket.IsObjectExist(filePath)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	return result
}

// HasDir 判断目录是否存在
func (storage *AliOssStorage) HasDir(dirPath string) bool {
	return true
}

// Read 读取文件内容
func (storage *AliOssStorage) Read(filePath string) ([]byte, error) {

	body, err := storage.bucket.GetObject(filePath)
	if err != nil {
		log.Println(err.Error())
		return []byte{}, err
	}
	defer func() {
		_ = body.Close()
	}()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		log.Println(err.Error())
		return []byte{}, err
	}

	return data, nil
}

// ReadStream 读取文件流
func (storage *AliOssStorage) ReadStream(filePath string, mode string) (io.ReadCloser, error) {
	body, err := storage.bucket.GetObject(filePath)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer func() {
		_ = body.Close()
	}()

	return body, nil
}

// Save 保存文件
func (storage *AliOssStorage) Save(dstFile string, srcFile io.Reader, mime string) (bool, error) {

	var options []oss.Option
	if mime != "" {
		options = []oss.Option{
			oss.ContentType(mime),
		}
	}

	if err := storage.bucket.PutObject(dstFile, srcFile, options...); err != nil {
		return false, err
	}

	return true, nil
}

// Cover 生成缩略图封面
func (storage *AliOssStorage) Cover(sourceImagePath, coverImagePath string, width, height uint) error {
	style := "image/resize,m_lfit"
	if width > 0 {
		style += ",w_" + strconv.Itoa(int(width))
	}
	if height > 0 {
		style += ",h_" + strconv.Itoa(int(height))
	}
	process := fmt.Sprintf("%s|sys/saveas,o_%v", style, base64.URLEncoding.EncodeToString([]byte(coverImagePath)))

	_, err := storage.bucket.ProcessObject(sourceImagePath, process)
	if err != nil {
		log.Println(err)
		return errors.New("缩略图生成失败")
	}

	return nil
}

// Delete 删除文件
func (storage *AliOssStorage) Delete(filePath string) (bool, error) {

	if err := storage.bucket.DeleteObject(filePath); err != nil {
		log.Println(err.Error())
		return false, err
	}

	return true, nil
}

// MultipleDelete 删除多个文件
func (storage *AliOssStorage) MultipleDelete(filePath []string) (bool, error) {

	_, err := storage.bucket.DeleteObjects(filePath)
	oss.DeleteObjectsQuiet(true)
	if err != nil {
		log.Println(err)
		return false, err
	}

	return true, nil
}

// MkDir 创建目录
func (storage *AliOssStorage) MkDir(dir string, mode os.FileMode) (bool, error) {
	return true, nil
}

// SignUrl 获取授权资源路径
func (storage *AliOssStorage) SignUrl(object string) string {

	signUrl, err := storage.bucket.SignURL(object, oss.HTTPGet, 60 + 8 * 3600)
	if err != nil {
		log.Println(err.Error())
		return object
	}

	return strings.Replace(signUrl, "http://", "https://", 1)
}

// FullUrl 完整路径
func (storage *AliOssStorage) FullUrl(object string) string {
	return "https://" + storage.bucketName + "." + storage.endpoint + "/" + object
}

// OriginalObject 获取原始资源路径
func (storage *AliOssStorage) OriginalObject(object string) string {
	u, err := url.Parse(object)
	if err != nil {
		return object
	}
	return strings.TrimLeft(u.Path, "/")
}
