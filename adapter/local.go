package adapter

import (
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/dysodeng/filesystem/storage"
)

// LocalAdapter 本地文件存储适配器
type LocalAdapter struct {
	config LocalConfig
}

type LocalConfig struct {
	BasePath string
	BaseUrl  string
	UseSSL   bool
}

func NewLocalAdapter(config LocalConfig) Adapter {
	config.BasePath = strings.TrimRight(config.BasePath, "/") + "/"
	return &LocalAdapter{
		config: config,
	}
}

// absolutePath 文件绝对路径
func (adapter *LocalAdapter) absolutePath(filename string) string {
	return adapter.config.BasePath + strings.TrimLeft(filename, "/")
}

// isReadable 是否有可读权限
func (adapter *LocalAdapter) isReadable(filename string) bool {
	err := syscall.Access(adapter.absolutePath(filename), syscall.O_RDONLY)
	if err != nil {
		return false
	}
	return true
}

// isWritable 是否有可写权限
func (adapter *LocalAdapter) isWritable(filename string) bool {
	err := syscall.Access(adapter.absolutePath(filename), syscall.O_RDWR)
	if err != nil {
		return false
	}
	return true
}

func (adapter *LocalAdapter) Info(file string) (storage.Attribute, error) {
	info, err := os.Stat(adapter.absolutePath(file))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, FileNotExists
		}
		return nil, err
	}

	if !adapter.isReadable(file) {
		return nil, FileNotReadable
	}

	if info.IsDir() {
		return storage.NewDirectoryAttribute(file, "", info.ModTime().Unix()), nil
	}

	// mime type
	f, _ := os.Open(adapter.absolutePath(file))

	buffer := make([]byte, 512)
	_, _ = f.Read(buffer)

	contentType := http.DetectContentType(buffer)

	return storage.NewFileAttribute(file, "", contentType, info.Size(), info.ModTime().Unix()), nil
}

func (adapter *LocalAdapter) HasFile(file string) bool {
	info, err := os.Stat(adapter.absolutePath(file))
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	return true
}

func (adapter *LocalAdapter) HasDir(file string) bool {
	info, err := os.Stat(adapter.absolutePath(file))
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	return true
}

func (adapter *LocalAdapter) Read(file string) (io.ReadCloser, error) {
	if !adapter.HasFile(file) {
		return nil, FileNotExists
	}

	if !adapter.isReadable(file) {
		return nil, FileNotReadable
	}

	f, err := os.Open(adapter.absolutePath(file))
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (adapter *LocalAdapter) Save(dstFile string, srcFile io.Reader, mimeType string) (bool, error) {
	dst, err := os.OpenFile(adapter.absolutePath(dstFile), os.O_WRONLY|os.O_CREATE, os.FileMode(0644))
	if err != nil {
		return false, err
	}
	defer func() {
		_ = dst.Close()
	}()

	_, err = io.Copy(dst, srcFile)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (adapter *LocalAdapter) Cover(sourceImagePath, coverImagePath string, width, height uint) error {
	// TODO implement me
	panic("implement me")
}

func (adapter *LocalAdapter) Copy(srcFile, disFile string) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (adapter *LocalAdapter) Move(disFile, srcFile string) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (adapter *LocalAdapter) Delete(file string) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (adapter *LocalAdapter) MultipleDelete(fileList []string) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (adapter *LocalAdapter) MkDir(dir string, mode os.FileMode) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (adapter *LocalAdapter) DeleteDir(dir string) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (adapter *LocalAdapter) List(dir string, iterable func(attribute storage.Attribute)) error {
	// TODO implement me
	panic("implement me")
}

func (adapter *LocalAdapter) FullPath(path string) string {
	var s = "http"
	if adapter.config.UseSSL {
		s = "https"
	}
	return s + "://" + adapter.config.BaseUrl + "/" + strings.TrimLeft(path, "/")
}

func (adapter *LocalAdapter) OriginalPath(fullPath string) string {
	return strings.TrimLeft(fullPath, adapter.config.BaseUrl)
}
