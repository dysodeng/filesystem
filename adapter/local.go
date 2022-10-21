package adapter

import (
	"github.com/dysodeng/filesystem/storage"
	"io"
	"os"
)

type LocalAdapter struct{}

func (adapter *LocalAdapter) Info(file string) (storage.Attribute, error) {
	return nil, nil
}

func (adapter *LocalAdapter) HasFile(file string) bool {
	return false
}

func (adapter *LocalAdapter) HasDir(file string) bool {
	return false
}

func (adapter *LocalAdapter) Read(file string) ([]byte, error) {
	return nil, nil
}

func (adapter *LocalAdapter) ReadStream(file string, mode string) (io.ReadCloser, error) {
	return nil, nil
}

func (adapter *LocalAdapter) Save(dstFile string, srcFile io.Reader, mimeType string) (bool, error) {
	return false, nil
}

func (adapter *LocalAdapter) Copy(disFile, srcFile string) (bool, error) {
	return false, nil
}

func (adapter *LocalAdapter) Move(disFile, srcFile string) (bool, error) {
	return false, nil
}

func (adapter *LocalAdapter) Cover(sourceImagePath, coverImagePath string, width, height uint) error {
	return nil
}

func (adapter *LocalAdapter) Delete(filePath string) (bool, error) {
	return false, nil
}

func (adapter *LocalAdapter) MultipleDelete(filePath []string) (bool, error) {
	return false, nil
}

func (adapter *LocalAdapter) MkDir(dir string, mode os.FileMode) (bool, error) {
	return false, nil
}

func (adapter *LocalAdapter) DeleteDir(dir string) (bool, error) {
	return false, nil
}

func (adapter *LocalAdapter) List(dir string, iterable func(attribute storage.Attribute)) error {
	return nil
}

func (adapter *LocalAdapter) FullPath(path string) string {
	return ""
}

func (adapter *LocalAdapter) OriginalPath(fullPath string) string {
	return ""
}
