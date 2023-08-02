package storage

import (
	"encoding/json"
	"errors"
	"strings"
)

type (
	FileType   string // 文件类型
	Visibility string // 文件可见性

	Attribute interface {
		Name() string
		Path() string
		Type() FileType
		Visibility() Visibility
		LastModified() int64

		IsFile() bool
		IsDir() bool

		MarshalJSON() ([]byte, error)
		UnmarshalJSON([]byte) error
	}
)

const (
	File      FileType = "file"
	Directory FileType = "directory"
)

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

type FileAttribute struct {
	name         string
	path         string
	fileSize     int64
	lastModified int64
	visibility   Visibility
	mimeType     string
}

type jsonAttr struct {
	Name         string     `json:"name"`
	Path         string     `json:"path"`
	Type         FileType   `json:"type"`
	FileSize     int64      `json:"file_size"`
	LastModified int64      `json:"last_modified"`
	Visibility   Visibility `json:"visibility"`
	MimeType     string     `json:"mime_type"`
}

func NewFileAttribute(name, path string, visibility Visibility, mimeType string, fileSize, lastModified int64) *FileAttribute {
	if visibility == "" {
		visibility = VisibilityPublic
	}
	return &FileAttribute{
		name:         name,
		path:         path,
		visibility:   visibility,
		mimeType:     mimeType,
		fileSize:     fileSize,
		lastModified: lastModified,
	}
}

func (file *FileAttribute) Name() string {
	return file.name
}

func (file *FileAttribute) Path() string {
	return file.path
}

func (file *FileAttribute) Type() FileType {
	return File
}

func (file *FileAttribute) FileSize() int64 {
	return file.fileSize
}

func (file *FileAttribute) Visibility() Visibility {
	return file.visibility
}

func (file *FileAttribute) LastModified() int64 {
	return file.lastModified
}

func (file *FileAttribute) MimeType() string {
	return file.mimeType
}

func (file *FileAttribute) IsFile() bool {
	return true
}

func (file *FileAttribute) IsDir() bool {
	return false
}

func (file *FileAttribute) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":          file.name,
		"path":          file.path,
		"type":          file.Type(),
		"file_size":     file.fileSize,
		"last_modified": file.lastModified,
		"visibility":    file.visibility,
		"mime_type":     file.mimeType,
	})
}

func (file *FileAttribute) UnmarshalJSON(data []byte) error {
	var m jsonAttr
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	if m.Type != File {
		return errors.New("类型错误")
	}

	file.name = m.Name
	file.path = m.Path
	file.fileSize = m.FileSize
	file.lastModified = m.LastModified
	file.visibility = m.Visibility
	file.mimeType = m.MimeType
	return nil
}

type DirectoryAttribute struct {
	name         string
	path         string
	lastModified int64
	visibility   Visibility
}

func NewDirectoryAttribute(name, path string, visibility Visibility, lastModified int64) *DirectoryAttribute {
	if visibility == "" {
		visibility = VisibilityPublic
	}
	return &DirectoryAttribute{
		name:         name,
		path:         strings.TrimRight(path, "/"),
		visibility:   visibility,
		lastModified: lastModified,
	}
}

func (dir *DirectoryAttribute) Name() string {
	return dir.name
}

func (dir *DirectoryAttribute) Path() string {
	return dir.path
}

func (dir *DirectoryAttribute) Type() FileType {
	return Directory
}

func (dir *DirectoryAttribute) Visibility() Visibility {
	return dir.visibility
}

func (dir *DirectoryAttribute) LastModified() int64 {
	return dir.lastModified
}

func (dir *DirectoryAttribute) IsFile() bool {
	return false
}

func (dir *DirectoryAttribute) IsDir() bool {
	return true
}

func (dir *DirectoryAttribute) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":          dir.name,
		"path":          dir.path,
		"type":          dir.Type(),
		"last_modified": dir.lastModified,
		"visibility":    dir.visibility,
	})
}

func (dir *DirectoryAttribute) UnmarshalJSON(data []byte) error {
	var m jsonAttr
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	if m.Type != Directory {
		return errors.New("类型错误")
	}

	dir.name = m.Name
	dir.path = m.Path
	dir.lastModified = m.LastModified
	dir.visibility = m.Visibility
	return nil
}
