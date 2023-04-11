package adapter

import (
	"io"
	"os"

	"github.com/dysodeng/filesystem/storage"
	"github.com/pkg/errors"
)

var (
	FileNotExists        = errors.New("file or directory does not exists")
	FileNotReadable      = errors.New("file is not readable")
	FileNotWritable      = errors.New("file is not writable")
	DirectoryNotWritable = errors.New("directory is not writable")
)

// Adapter 存储适配器接口
type Adapter interface {
	// Info 文件/目录信息
	// @param file string 文件路径
	Info(file string) (storage.Attribute, error)

	// HasFile 判断文件是否存在
	// @param file string 文件路径
	HasFile(file string) bool

	// HasDir 判断目录是否存在
	// @param file string 文件路径
	HasDir(file string) bool

	// Read 读取文件内容
	// @param file string 文件路径
	Read(file string) (io.ReadCloser, error)

	// Save 保存文件
	// @param dstFile string 目标文件路径
	// @param srcFile io.Reader 原文件内容
	Save(dstFile string, srcFile io.Reader, mimeType string) (bool, error)

	// Cover 生成缩略图封面
	// @param sourceImagePath string 原文件路径
	// @param coverImagePath string 目标文件路径
	Cover(sourceImagePath, coverImagePath string, width, height uint) error

	// Copy 复制文件/目录
	// @param srcFile string 原文件路径
	// @param dstFile string 目标文件路径
	Copy(srcFile, disFile string) (bool, error)

	// Move 移动文件/目录
	// @param dstFile string 目标文件路径
	// @param srcFile string 原文件路径
	Move(dstFile, srcFile string) (bool, error)
  
	// Delete 删除文件
	// @param file string 文件路径
	Delete(file string) (bool, error)

	// MultipleDelete 删除多个文件
	// @param fileList []string 文件列表
	MultipleDelete(fileList []string) (bool, error)

	// MkDir 创建目录
	// @param dir string 目录路径
	MkDir(dir string, mode os.FileMode) (bool, error)

	// DeleteDir 删除目录
	// @param dir string 目录路径
	DeleteDir(dir string) (bool, error)

	// List 文件/目录列表
	// @param dir string 目录路径
	// @param iterable func 迭代器
	List(dir string, iterable func(attribute storage.Attribute)) error

	// FullPath 获取全路径
	// @param path string 文件路径
	FullPath(path string) string

	// OriginalPath 获取原始路径
	// @param fullPath string 文件全路径
	OriginalPath(fullPath string) string
}
