package adapter

import (
	"fmt"
	"github.com/dysodeng/filesystem/storage"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
)

// HwObsAdapter 华为云OBS存储适配器
type HwObsAdapter struct {
	client *obs.ObsClient
	config HwObsConfig
}

type HwObsConfig struct {
	AccessKey      string
	SecretKey      string
	EndPoint       string
	BucketName     string
	StayBucketName string
	IsPrivate      bool
}

func NewHwObsAdapter(config HwObsConfig) Adapter {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	hwObsAdapter := new(HwObsAdapter)

	client, err := obs.New(config.AccessKey, config.SecretKey, config.EndPoint)
	if err != nil {
		panic("hw_obs connect error:" + err.Error())
	}

	hwObsAdapter.client = client
	hwObsAdapter.config = config

	return hwObsAdapter
}

func (adapter *HwObsAdapter) Info(file string) (storage.Attribute, error) {
	input := &obs.GetObjectMetadataInput{
		Bucket: adapter.config.BucketName,
		Key:    file,
	}
	output, err := adapter.client.GetObjectMetadata(input)
	if err != nil {
		return nil, FileNotExists
	}

	return storage.NewFileAttribute(file, "", output.ContentType, output.ContentLength, output.LastModified.Unix()), nil
}

func (adapter *HwObsAdapter) HasFile(file string) bool {
	_, err := adapter.Info(file)
	if err != nil {
		return false
	}
	return true
}

func (adapter *HwObsAdapter) HasDir(file string) bool {
	return true
}

func (adapter *HwObsAdapter) Read(file string) (io.ReadCloser, error) {
	input := &obs.GetObjectInput{}
	input.Bucket = adapter.config.BucketName
	input.Key = file

	output, err := adapter.client.GetObject(input)
	if err != nil {
		return nil, err
	}

	return output.Body, nil
}

func (adapter *HwObsAdapter) Save(dstFile string, srcFile io.Reader, mimeType string) (bool, error) {
	input := &obs.PutObjectInput{
		Body: srcFile,
	}
	input.Bucket = adapter.config.BucketName
	input.Key = dstFile
	if mimeType != "" {
		input.ContentType = mimeType
	}

	_, err := adapter.client.PutObject(input)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (adapter *HwObsAdapter) Copy(srcFile, disFile string) (bool, error) {
	input := &obs.CopyObjectInput{}
	input.Bucket = adapter.config.BucketName
	input.Key = disFile
	input.CopySourceBucket = adapter.config.BucketName
	input.CopySourceKey = srcFile

	_, err := adapter.client.CopyObject(input)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (adapter *HwObsAdapter) Move(disFile, srcFile string) (bool, error) {
	if _, err := adapter.Copy(srcFile, disFile); err != nil {
		return false, err
	}
	if _, err := adapter.Delete(srcFile); err != nil {
		return false, err
	}
	return true, nil
}

func (adapter *HwObsAdapter) Cover(sourceImagePath, coverImagePath string, width, height uint) error {
	style := "image/resize,m_lfit"
	if width > 0 {
		style += fmt.Sprintf(",w_%d", width)
	}
	if height > 0 {
		style += fmt.Sprintf(",h_%d", height)
	}

	input := &obs.GetObjectInput{}
	input.Bucket = adapter.config.BucketName
	input.Key = sourceImagePath
	input.ImageProcess = style
	output, err := adapter.client.GetObject(input)
	if err != nil {
		return err
	}

	putInput := &obs.PutObjectInput{
		Body: output.Body,
	}
	putInput.Bucket = adapter.config.BucketName
	putInput.Key = coverImagePath
	_, err = adapter.client.PutObject(putInput)
	if err != nil {
		return err
	}

	return nil
}

func (adapter *HwObsAdapter) Delete(file string) (bool, error) {
	input := &obs.DeleteObjectInput{}
	input.Bucket = adapter.config.BucketName
	input.Key = file

	_, err := adapter.client.DeleteObject(input)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (adapter *HwObsAdapter) MultipleDelete(fileList []string) (bool, error) {
	length := len(fileList)
	if length <= 0 {
		return true, nil
	}

	input := &obs.DeleteObjectsInput{
		Bucket: adapter.config.BucketName,
	}
	objects := make([]obs.ObjectToDelete, length)
	for i := 0; i < length; i++ {
		objects[i].Key = fileList[i]
	}
	input.Objects = objects

	_, err := adapter.client.DeleteObjects(input)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (adapter *HwObsAdapter) MkDir(dir string, mode os.FileMode) (bool, error) {
	return true, nil
}

func (adapter *HwObsAdapter) DeleteDir(dir string) (bool, error) {
	return true, nil
}

func (adapter *HwObsAdapter) List(dir string, iterable func(attribute storage.Attribute)) error {
	dir = strings.TrimRight(dir, "/") + "/"
	input := &obs.ListObjectsInput{
		Bucket: adapter.config.BucketName,
	}
	if dir == "/" {
		dir = ""
	}

	prefixDir := dir

	for {
		input.Prefix = prefixDir
		output, err := adapter.client.ListObjects(input)
		if err != nil {
			return err
		}

		for _, prefix := range output.CommonPrefixes {
			iterable(storage.NewDirectoryAttribute(prefix, "", 0))
		}
		for _, content := range output.Contents {
			iterable(storage.NewFileAttribute(content.Key, "", "", content.Size, content.LastModified.Unix()))
		}

		if output.IsTruncated {
			prefixDir = output.Prefix
		} else {
			break
		}
	}

	return nil
}

func (adapter *HwObsAdapter) FullPath(path string) string {
	if adapter.config.IsPrivate {
		input := &obs.CreateSignedUrlInput{
			Method:  obs.HttpMethodGet,
			Bucket:  adapter.config.BucketName,
			Key:     path,
			Expires: 60 + 8*3600,
		}
		output, err := adapter.client.CreateSignedUrl(input)
		if err != nil {
			return ""
		}
		return strings.Replace(output.SignedUrl, "http://", "https://", 1)
	}
	return "https://" + adapter.config.BucketName + "." + adapter.config.EndPoint + "/" + path
}

func (adapter *HwObsAdapter) OriginalPath(fullPath string) string {
	u, err := url.Parse(fullPath)
	if err != nil {
		return fullPath
	}
	return strings.TrimLeft(u.Path, "/")
}
