package webdav

import (
	"errors"
	"lsgooi/types"
	"os"
	"syscall"
	"time"
)

// File is the file for a gooi file.
type File struct {
	*FileInfo
	file *os.File
}

func (f *File) Chdir() error                                  { return errors.New("not supported") }
func (f *File) Chmod(mode os.FileMode) error                  { return errors.New("not supported") }
func (f *File) Chown(uid, gid int) error                      { return errors.New("not supported") }
func (f *File) Close() error                                  { return f.file.Close() }
func (f *File) Fd() uintptr                                   { return f.file.Fd() }
func (f *File) Name() string                                  { return f.fname }
func (f *File) Read(b []byte) (n int, err error)              { return f.file.Read(b) }
func (f *File) ReadAt(b []byte, off int64) (n int, err error) { return f.file.ReadAt(b, off) }
func (f *File) Readdirnames(n int) (names []string, err error) {
	items, err := f.Readdir(n)
	if err != nil {
		return []string{}, err
	}

	var res []string
	for _, item := range items {
		res = append(res, item.Name())
	}
	return res, nil
}
func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	return f.file.Seek(offset, whence)
}
func (f *File) SetDeadline(t time.Time) error                  { return errors.New("not supported") }
func (f *File) SetReadDeadline(t time.Time) error              { return errors.New("not supported") }
func (f *File) SetWriteDeadline(t time.Time) error             { return errors.New("not supported") }
func (f *File) Stat() (os.FileInfo, error)                     { return f, nil }
func (f *File) Sync() error                                    { return errors.New("not supported") }
func (f *File) SyscallConn() (syscall.RawConn, error)          { return nil, errors.New("not supported") }
func (f *File) Truncate(size int64) error                      { return errors.New("not supported") }
func (f *File) Write(b []byte) (n int, err error)              { return 0, errors.New("not supported") }
func (f *File) WriteAt(b []byte, off int64) (n int, err error) { return 0, errors.New("not supported") }
func (f *File) WriteString(s string) (n int, err error)        { return 0, errors.New("not supported") }

func MakeFile(fs *FileSystem, realPath, fname string, item types.Item) (*File, error) {
	f := &File{MakeFileInfo(fs, fname, item), nil}

	if realPath != "" {
		realFile, err := os.Open(realPath)
		if err != nil {
			return nil, err
		}
		f.file = realFile
	}

	return f, nil
}
