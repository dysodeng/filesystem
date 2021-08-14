package storage

import (
	"io"
	"os"
)

// Storage 文件存储器接口
type Storage interface {

	// HasFile 判断文件是否存在
	HasFile(filePath string) bool

	// HasDir 判断目录是否存在
	HasDir(dirPath string) bool

	// Read 读取文件内容
	Read(filePath string) ([]byte, error)

	// ReadStream 读取文件流
	ReadStream(filePath string, mode string) (io.ReadCloser, error)

	// Save 保存文件
	Save(dstFile string, srcFile io.Reader, mime string) (bool, error)

	// Cover 生成缩略图封面
	Cover(sourceImagePath, coverImagePath string, width, height uint) error

	// Delete 删除文件
	Delete(filePath string) (bool, error)

	// MultipleDelete 删除多个文件
	MultipleDelete(filePath []string) (bool, error)

	// MkDir 创建目录
	MkDir(dir string, mode os.FileMode) (bool, error)

	// SignUrl 获取授权资源路径
	SignUrl(object string) string

	// FullUrl 完整路径
	FullUrl(object string) string

	// OriginalObject 获取原始资源路径
	OriginalObject(object string) string
}
