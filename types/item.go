package types

import (
	"time"

	"github.com/dustin/go-humanize"
)

// Item represents an gooid item on disk.
type Item struct {
	ID   string
	Name string
	Size uint64
	Date time.Time
	URL  string
}

// SizeString returns the size of the current item in a human friendly format.
func (item Item) SizeString() string {
	return humanize.Bytes(item.Size)
}

// DateString returns the modification date of the current item in a human and
// machine friendly format.
func (item Item) DateString() string {
	return item.Date.Format("2006-01-02 15:04:05")
}
