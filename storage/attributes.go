package storage

import (
	"encoding/json"
	"errors"
	"strings"
)

const (
	File      string = "file"
	Directory string = "directory"
)

const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
)

type Attribute interface {
	Path() string
	Type() string
	Visibility() string
	LastModified() int64

	IsFile() bool
	IsDir() bool

	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
}

type FileAttribute struct {
	path         string
	fileSize     int64
	lastModified int64
	visibility   string
	mimeType     string
}

type jsonAttr struct {
	Path         string `json:"path"`
	Type         string `json:"type"`
	FileSize     int64  `json:"file_size"`
	LastModified int64  `json:"last_modified"`
	Visibility   string `json:"visibility"`
	MimeType     string `json:"mime_type"`
}

func NewFileAttribute(path, visibility, mimeType string, fileSize, lastModified int64) *FileAttribute {
	if visibility == "" {
		visibility = VisibilityPublic
	}
	return &FileAttribute{
		path:         path,
		visibility:   visibility,
		mimeType:     mimeType,
		fileSize:     fileSize,
		lastModified: lastModified,
	}
}

func (file *FileAttribute) Path() string {
	return file.path
}

func (file *FileAttribute) Type() string {
	return File
}

func (file *FileAttribute) FileSize() int64 {
	return file.fileSize
}

func (file *FileAttribute) Visibility() string {
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

	file.path = m.Path
	file.fileSize = m.FileSize
	file.lastModified = m.LastModified
	file.visibility = m.Visibility
	file.mimeType = m.MimeType
	return nil
}

type DirectoryAttribute struct {
	path         string
	lastModified int64
	visibility   string
}

func NewDirectoryAttribute(path, visibility string, lastModified int64) *DirectoryAttribute {
	if visibility == "" {
		visibility = VisibilityPublic
	}
	return &DirectoryAttribute{
		path:         strings.Trim(path, "/"),
		visibility:   visibility,
		lastModified: lastModified,
	}
}

func (dir *DirectoryAttribute) Path() string {
	return dir.path
}

func (dir *DirectoryAttribute) Type() string {
	return Directory
}

func (dir *DirectoryAttribute) Visibility() string {
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

	dir.path = m.Path
	dir.lastModified = m.LastModified
	dir.visibility = m.Visibility
	return nil
}
