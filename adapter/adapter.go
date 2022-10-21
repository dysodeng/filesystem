package adapter

import (
	"github.com/dysodeng/filesystem/storage"
	"github.com/pkg/errors"
	"io"
	"os"
)

var (
	FileNotExists = errors.New("file or directory does not exists")
)

// Adapter 存储适配器接口
type Adapter interface {
	// Info 文件/目录信息
	Info(file string) (storage.Attribute, error)

	// HasFile 判断文件是否存在
	HasFile(file string) bool

	// HasDir 判断目录是否存在
	HasDir(file string) bool

	// Read 读取文件内容
	Read(file string) ([]byte, error)

	// ReadStream 读取文件流
	ReadStream(file string, mode string) (io.ReadCloser, error)

	// Save 保存文件
	Save(dstFile string, srcFile io.Reader, mimeType string) (bool, error)

	// Copy 复制文件/目录
	Copy(srcFile, disFile string) (bool, error)

	// Move 移动文件/目录
	Move(disFile, srcFile string) (bool, error)

	// Cover 生成缩略图封面
	Cover(sourceImagePath, coverImagePath string, width, height uint) error

	// Delete 删除文件
	Delete(filePath string) (bool, error)

	// MultipleDelete 删除多个文件
	MultipleDelete(filePath []string) (bool, error)

	// MkDir 创建目录
	MkDir(dir string, mode os.FileMode) (bool, error)

	// DeleteDir 删除目录
	DeleteDir(dir string) (bool, error)

	// List 文件/目录列表
	List(dir string, iterable func(attribute storage.Attribute)) error

	// FullPath 获取全路径
	FullPath(path string) string

	// OriginalPath 获取原始路径
	OriginalPath(fullPath string) string
}
