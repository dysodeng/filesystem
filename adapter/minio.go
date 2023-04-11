package adapter

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/dysodeng/filesystem/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioAdapter Minio存储适配器,兼容AWS S3
type MinioAdapter struct {
	client *minio.Client
	config MinioConfig
}

type MinioConfig struct {
	AccessKey  string
	SecretKey  string
	EndPoint   string
	BucketName string
	UseSSL     bool // 是否使用https
	IsPrivate  bool // 是否私有访问权限
	IsAwsS3    bool // 是否为AWS S3存储
}

func NewMinioAdapter(config MinioConfig) Adapter {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	minioAdapter := new(MinioAdapter)

	client, err := minio.New(config.EndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		panic("hw_obs connect error:" + err.Error())
	}

	minioAdapter.client = client
	minioAdapter.config = config

	return minioAdapter
}

func (adapter *MinioAdapter) Info(file string) (storage.Attribute, error) {
	info, err := adapter.client.StatObject(context.Background(), adapter.config.BucketName, file, minio.StatObjectOptions{})
	if err != nil {
		return nil, FileNotExists
	}

	return storage.NewFileAttribute(info.Key, "", info.ContentType, info.Size, info.LastModified.Unix()), nil
}

func (adapter *MinioAdapter) HasFile(file string) bool {
	_, err := adapter.Info(file)
	if err != nil {
		return false
	}
	return true
}

func (adapter *MinioAdapter) HasDir(file string) bool {
	return true
}

func (adapter *MinioAdapter) Read(file string) (io.ReadCloser, error) {
	object, err := adapter.client.GetObject(context.Background(), adapter.config.BucketName, file, minio.GetObjectOptions{})
	if err != nil {
		return nil, FileNotExists
	}
	return object, nil
}

func (adapter *MinioAdapter) Save(dstFile string, srcFile io.Reader, mimeType string) (bool, error) {
	content, err := io.ReadAll(srcFile)
	if err != nil {
		return false, err
	}
	_, err = adapter.client.PutObject(
		context.Background(),
		adapter.config.BucketName,
		dstFile,
		bytes.NewReader(content),
		int64(len(content)),
		minio.PutObjectOptions{ContentType: mimeType},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (adapter *MinioAdapter) Cover(sourceImagePath, coverImagePath string, width, height uint) error {
	sourceImageBytes, err := os.ReadFile(sourceImagePath)
	if err != nil {
		return err
	}

	format, _ := imaging.FormatFromFilename(sourceImagePath)

	sourceImage, _, err := image.Decode(bytes.NewReader(sourceImageBytes))
	if err != nil {
		return err
	}

	coverImage := imaging.Resize(sourceImage, int(width), int(height), imaging.Lanczos)
	writer := bytes.NewBuffer(nil)

	err = imaging.Encode(writer, coverImage, format)
	if err != nil {
		return err
	}

	_, err = adapter.client.PutObject(
		context.Background(),
		adapter.config.BucketName,
		coverImagePath,
		writer,
		int64(writer.Len()),
		minio.PutObjectOptions{ContentType: format.String()},
	)
	if err != nil {
		return err
	}

	return nil
}

func (adapter *MinioAdapter) Copy(srcFile, dstFile string) (bool, error) {
	src := minio.CopySrcOptions{
		Bucket: adapter.config.BucketName,
		Object: srcFile,
	}

	dst := minio.CopyDestOptions{
		Bucket: adapter.config.BucketName,
		Object: dstFile,
	}

	_, err := adapter.client.CopyObject(context.Background(), dst, src)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (adapter *MinioAdapter) Move(dstFile, srcFile string) (bool, error) {
	_, err := adapter.Copy(srcFile, dstFile)
	if err != nil {
		return false, err
	}

	_, err = adapter.Delete(srcFile)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (adapter *MinioAdapter) Delete(file string) (bool, error) {
	err := adapter.client.RemoveObject(context.Background(), adapter.config.BucketName, file, minio.RemoveObjectOptions{GovernanceBypass: true})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (adapter *MinioAdapter) MultipleDelete(fileList []string) (bool, error) {
	for _, file := range fileList {
		err := adapter.client.RemoveObject(context.Background(), adapter.config.BucketName, file, minio.RemoveObjectOptions{GovernanceBypass: true})
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (adapter *MinioAdapter) MkDir(dir string, mode os.FileMode) (bool, error) {
	return true, nil
}

func (adapter *MinioAdapter) DeleteDir(dir string) (bool, error) {
	return true, nil
}

func (adapter *MinioAdapter) List(dir string, iterable func(attribute storage.Attribute)) error {
	dir = strings.TrimRight(dir, "/") + "/"
	if dir == "/" {
		dir = ""
	}

	opts := minio.ListObjectsOptions{
		Recursive: false,
		Prefix:    dir,
	}

	for object := range adapter.client.ListObjects(context.Background(), adapter.config.BucketName, opts) {
		if object.Err != nil {
			fmt.Println(object.Err)
			continue
		}

		if string(object.Key[len(object.Key)-1]) == "/" {
			iterable(storage.NewDirectoryAttribute(object.Key, "", 0))
		} else {
			iterable(storage.NewFileAttribute(object.Key, "", object.ContentType, object.Size, object.LastModified.Unix()))
		}
	}

	return nil
}

func (adapter *MinioAdapter) FullPath(path string) string {
	if adapter.config.IsPrivate {
		signUrl, err := adapter.client.PresignedGetObject(
			context.Background(),
			adapter.config.BucketName,
			path,
			time.Hour*3,
			nil,
		)
		if err != nil {
			return ""
		}
		if adapter.config.UseSSL {
			return strings.Replace(signUrl.String(), "http://", "https://", 1)
		}
		return signUrl.String()
	}

	var urlBuilder strings.Builder
	if adapter.config.UseSSL {
		urlBuilder.WriteString("https")
	} else {
		urlBuilder.WriteString("http")
	}
	urlBuilder.WriteString("://")
	if adapter.config.IsAwsS3 {
		urlBuilder.WriteString(adapter.config.BucketName)
		urlBuilder.WriteString(".")
		urlBuilder.WriteString(adapter.config.EndPoint)
		urlBuilder.WriteString("/")
		urlBuilder.WriteString(path)
	} else {
		urlBuilder.WriteString(adapter.config.EndPoint)
		urlBuilder.WriteString("/")
		urlBuilder.WriteString(adapter.config.BucketName)
		urlBuilder.WriteString("/")
		urlBuilder.WriteString(path)
	}

	return urlBuilder.String()
}

func (adapter *MinioAdapter) OriginalPath(fullPath string) string {
	u, err := url.Parse(fullPath)
	if err != nil {
		return fullPath
	}
	return strings.TrimLeft(u.Path, "/")
}
