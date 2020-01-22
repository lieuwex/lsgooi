package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"lsgooi/types"
	"path"
	"strings"
	"time"
)

// State contains all the state of the program
type State struct {
	// tpl contains the current compiled version of the page.
	tpl []byte
	// itemMap contains the currently known id to item mapping
	itemMap map[string]types.Item
	// lastCheckTime is the time the dir was last checked of new files
	lastCheckTime time.Time
}

// MustCheck returns whether or not to check of new dir items
func (s State) MustCheck() bool {
	return time.Now().Sub(s.lastCheckTime).Seconds() >= 10
}

// Update checks whether or not the template should be updated, and updates if
// necessary.
func (state *State) Update() {
	if !state.MustCheck() {
		return
	}

	itemMap, err := readItems(dir, state.itemMap)
	if err != nil {
		panic(err)
	}

	state.lastCheckTime = time.Now()

	oldLen := len(state.itemMap)
	newLen := len(itemMap)
	if oldLen != newLen {
		log.Printf("read %d new file(s)", newLen-oldLen)

		state.itemMap = itemMap
		state.tpl, err = compileTemplate(itemMap)
		if err != nil {
			panic(err)
		}
		webdavHandler.Refresh(state.itemMap)
	}
}

// readItems reads the given gooi files directory for items, if prev is non-nil
// it will be used as a cache for existing files. If a file is removed from the
// directory it isn't included in the result, even if prev does contain it.
func readItems(dir string, prev map[string]types.Item) (map[string]types.Item, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	m := make(map[string]types.Item)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "-fname") || f.Name() == "startid" {
			continue
		}

		id := f.Name()

		if val, has := prev[id]; has {
			m[id] = val
			continue
		}

		fname, err := ioutil.ReadFile(path.Join(dir, id+"-fname"))
		if err != nil {
			return m, err
		}

		m[id] = types.Item{
			ID:   id,
			Name: strings.TrimSpace(string(fname)),
			Size: uint64(f.Size()),
			Date: f.ModTime(),
			URL:  fmt.Sprintf(urlfmt, id, fname),
		}
	}
	return m, nil
}
