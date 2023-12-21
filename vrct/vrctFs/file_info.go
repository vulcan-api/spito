package vrctFs

import (
	"io/fs"
	"time"
)

type FileInfo struct {
	name     string
	size     int64
	fileMode fs.FileMode
	modTime  time.Time
	isDir    bool
}

func (f FileInfo) Name() string {
	return f.name
}

func (f FileInfo) Size() int64 {
	return f.size
}

func (f FileInfo) Mode() fs.FileMode {
	return f.fileMode
}

func (f FileInfo) ModTime() time.Time {
	return f.modTime
}

func (f FileInfo) IsDir() bool {
	return f.isDir
}

func (f FileInfo) Sys() any {
	// TODO: implement this function
	return nil
}
