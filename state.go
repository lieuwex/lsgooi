package main

import (
	"lsgooi/types"
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
