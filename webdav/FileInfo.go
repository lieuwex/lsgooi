package webdav

import (
	"lsgooi/types"
	"os"
	"time"
)

// FileInfo is the info for a gooi file
type FileInfo struct {
	// item is the underlying gooi item
	item types.Item
	// fname is the mapped file name
	fname string

	// is root tells whether or not this it he gooi root
	isRoot bool

	fs *FileSystem // HACK
}

func MakeFileInfo(fs *FileSystem, fname string, item types.Item) *FileInfo {
	return &FileInfo{item, fname, false, fs}
}

func (i *FileInfo) Name() string       { return i.fname }
func (i *FileInfo) Size() int64        { return int64(i.item.Size) }
func (i *FileInfo) Mode() os.FileMode  { return 0777 }
func (i *FileInfo) ModTime() time.Time { return i.item.Date }
func (i *FileInfo) IsDir() bool        { return i.isRoot }
func (i *FileInfo) Sys() interface{}   { return nil }

func (i *FileInfo) Readdir(max int) ([]os.FileInfo, error) {
	if !i.isRoot {
		return []os.FileInfo{}, os.ErrInvalid
	}

	i.fs.m.RLock()
	defer i.fs.m.RUnlock()

	res := make([]os.FileInfo, 0, len(i.fs.mapping))

	for _, item := range i.fs.mapping {
		res = append(res, item)

		if max > 0 && len(res) >= max {
			break
		}
	}

	return res, nil
}
