package adapter

import (
	"bytes"
	"image"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/disintegration/imaging"

	"github.com/dysodeng/filesystem/storage"
)

// LocalAdapter 本地文件存储适配器
type LocalAdapter struct {
	config LocalConfig
}

type LocalConfig struct {
	BasePath string
	BaseUrl  string
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
	log.Println(adapter.absolutePath(filename))
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
		return storage.NewDirectoryAttribute(info.Name(), file, "", info.ModTime().Unix()), nil
	}

	// mime type
	f, _ := os.Open(adapter.absolutePath(file))

	buffer := make([]byte, 512)
	_, _ = f.Read(buffer)

	contentType := http.DetectContentType(buffer)

	return storage.NewFileAttribute(info.Name(), file, "", contentType, info.Size(), info.ModTime().Unix()), nil
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
	sourceImagePath = adapter.absolutePath(sourceImagePath)

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

	if _, err = adapter.Save(coverImagePath, writer, format.String()); err != nil {
		return err
	}

	return nil
}

func (adapter *LocalAdapter) Copy(srcFile, dstFile string) (bool, error) {
	dst, err := os.OpenFile(adapter.absolutePath(dstFile), os.O_RDWR|os.O_CREATE, os.FileMode(0644))
	if err != nil {
		return false, err
	}
	defer func() {
		_ = dst.Close()
	}()

	src, err := os.Open(adapter.absolutePath(srcFile))
	if err != nil {
		return false, err
	}
	defer func() {
		_ = src.Close()
	}()

	_, err = io.Copy(dst, src)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (adapter *LocalAdapter) Move(dstFile, srcFile string) (bool, error) {
	if _, err := adapter.Copy(srcFile, dstFile); err != nil {
		return false, err
	}

	if _, err := adapter.Delete(srcFile); err != nil {
		return false, err
	}

	return true, nil
}

func (adapter *LocalAdapter) Delete(file string) (bool, error) {
	if !adapter.HasFile(file) {
		return false, FileNotExists
	}
	if !adapter.isWritable(filepath.Dir(file)) {
		return false, DirectoryNotWritable
	}

	if err := os.Remove(adapter.absolutePath(file)); err != nil {
		return false, err
	}

	return true, nil
}

func (adapter *LocalAdapter) MultipleDelete(fileList []string) (bool, error) {
	for _, file := range fileList {
		if _, err := adapter.Delete(file); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (adapter *LocalAdapter) MkDir(dir string, mode os.FileMode) (bool, error) {
	if !adapter.HasDir(dir) {
		if err := os.MkdirAll(adapter.absolutePath(dir), mode); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (adapter *LocalAdapter) DeleteDir(dir string) (bool, error) {
	if !adapter.HasDir(dir) {
		return false, FileNotExists
	}

	if err := os.Remove(adapter.absolutePath(dir)); err != nil {
		return false, err
	}

	return true, nil
}

func (adapter *LocalAdapter) List(dir string, iterable func(attribute storage.Attribute)) error {
	if !adapter.HasDir(dir) {
		return FileNotExists
	}

	err := filepath.Walk(adapter.absolutePath(dir), func(path string, info fs.FileInfo, err error) error {
		path = strings.Replace(path, adapter.config.BasePath, "", 1)
		var attribute storage.Attribute
		if info.IsDir() {
			attribute = storage.NewDirectoryAttribute(info.Name(), path, "", info.ModTime().Unix())
		} else {
			attribute = storage.NewFileAttribute(info.Name(), path, "", "", info.Size(), info.ModTime().Unix())
		}

		iterable(attribute)

		return nil
	})

	return err
}

func (adapter *LocalAdapter) FullPath(path string) string {
	var urlBuilder strings.Builder

	urlBuilder.WriteString(strings.TrimRight(adapter.config.BaseUrl, "/"))
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(strings.TrimLeft(path, "/"))

	return urlBuilder.String()
}

func (adapter *LocalAdapter) OriginalPath(fullPath string) string {
	u, err := url.Parse(fullPath)
	if err != nil {
		return fullPath
	}
	return strings.TrimLeft(u.Path, "/")
}
