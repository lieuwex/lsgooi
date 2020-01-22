package webdav

import (
	"context"
	"errors"
	"fmt"
	"lsgooi/types"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/net/webdav"
)

// FileSystem is the virtual gooi filesystem
type FileSystem struct {
	dir string

	m sync.RWMutex
	// mapping from file name to Item
	mapping map[string]*FileInfo
}

func (fs *FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	fs.m.RLock()
	defer fs.m.RUnlock()

	if name == "/" {
		finfo, err := MakeFile(fs, "", "/", types.Item{})
		if err != nil {
			return nil, err
		}
		finfo.isRoot = true
		return finfo, nil
	}

	item, has := fs.mapping[name[1:]]
	if !has {
		return nil, os.ErrNotExist
	}

	fullPath := path.Join(fs.dir, item.item.ID)
	return MakeFile(fs, fullPath, name[1:], item.item)
}
func (fs *FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	fs.m.RLock()
	defer fs.m.RUnlock()

	if name == "/" {
		finfo := MakeFileInfo(fs, "/", types.Item{})
		finfo.isRoot = true
		return finfo, nil
	}

	item, has := fs.mapping[name[1:]]
	if !has {
		return nil, os.ErrNotExist
	}

	return item, nil
}

func (fs *FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return errors.New("not supported")
}
func (fs *FileSystem) RemoveAll(ctx context.Context, name string) error {
	return errors.New("not supported")
}
func (fs *FileSystem) Rename(ctx context.Context, oldName, newName string) error {
	return errors.New("not supported")
}

type WebdavHandler struct {
	webdav.Handler
}

func (h *WebdavHandler) Refresh(itemMap map[string]types.Item) {
	fs := h.FileSystem.(*FileSystem)

	fs.m.Lock()
	defer fs.m.Unlock()

	items := make([]types.Item, 0, len(itemMap))
	for _, item := range itemMap {
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Date.Before(items[j].Date)
	})

	// mapping from file name to count
	counts := make(map[string]uint)

	fs.mapping = make(map[string]*FileInfo)
	for _, item := range items {
		count := counts[item.Name]
		counts[item.Name] = count + 1

		fname := item.Name
		if count > 0 {
			ext := filepath.Ext(fname)
			base := strings.TrimSuffix(fname, ext)
			fname = fmt.Sprintf("%s (%d)%s", base, count, ext)
		}

		fs.mapping[fname] = MakeFileInfo(fs, fname, item)
	}
}

func MakeHandler(dir string) *WebdavHandler {
	h := &WebdavHandler{}
	h.Prefix = "/webdav"

	fs := &FileSystem{}
	fs.dir = dir
	h.FileSystem = fs

	h.LockSystem = webdav.NewMemLS()
	return h
}
