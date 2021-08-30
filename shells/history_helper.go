package shells

import (
	"embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aardlabs/terminal-poc/app"
	"github.com/aardlabs/terminal-poc/historian"
	"github.com/aardlabs/terminal-poc/tools"

	"github.com/mitchellh/go-ps"
)

const ExpandPrev = "^^"

var ExpandChar = ExpandPrev[0]

func EachHistoryEntry(reverse bool, duration time.Duration, currentSession bool, cb func(*historian.Item) error) error {
	filter := &historian.Filter{
		Reverse: reverse,
	}

	if duration != 0 {
		at := time.Now().Add(-duration)
		if filter.Reverse {
			filter.Stop = &at
		} else {
			filter.Start = &at
		}
	}

	parent, err := GetParent()
	if err != nil {
		return err
	}

	parentPID := parent.Pid()

	var skip func(*historian.Item) bool
	if currentSession {
		skip = func(item *historian.Item) bool {
			return *item.ParentPID != parentPID
		}
	} else {
		skip = func(*historian.Item) bool { return false }
	}

	hist := openHistorian(parent.Executable(), true)
	defer hist.Close()

	return hist.Each(filter, func(item *historian.Item) error {
		if skip(item) {
			return nil
		}

		return cb(item)
	})
}

func GetHistoryEntry(index string) (*historian.Item, error) {
	parent, err := GetParent()
	if err != nil {
		return nil, err
	}

	hist := openHistorian(parent.Executable(), true)
	defer hist.Close()

	if index == ExpandPrev {
		return getMostRecentEntry(parent, hist, true, false)
	}

	idStr := index[1:]
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return nil, err
	}

	return getItem(hist, id)
}

func PutHistoryEntry(item *historian.Item) error {
	parent, err := GetParent()
	if err != nil {
		return err
	}

	hist := openHistorian(parent.Executable(), false)
	defer hist.Close()

	if item.ParentPID == nil {
		pp := parent.Pid()
		item.ParentPID = &pp
	}

	return hist.Put(item)
}

func GetHistoryLen() uint {
	parent, err := GetParent()
	if err != nil {
		panic(err)
	}

	hist := openHistorian(parent.Executable(), true)
	defer hist.Close()

	return hist.Len()
}

//--------------------------------------------------------------------------------

var historyDir = tools.MyPathTo("history")

//go:embed **/*
var thisDir embed.FS

func getMostRecentEntry(parent ps.Process, hist *historian.Historian, currentSession bool, incomplete bool) (*historian.Item, error) {
	var parentPID *int
	if currentSession {
		pid := parent.Pid()
		parentPID = &pid
	}

	// try twice to find a valid item (in case the most recent is our own call--not yet exited)
	for offset := int64(-1); offset > -3; offset-- {
		item, err := getRelativeItem(hist, parentPID, offset)
		if err != nil {
			return nil, err
		}

		if item != nil {
			if (incomplete && item.ExitStatus == nil) ||
				(!incomplete && item.ExitStatus != nil) {
				return item, nil
			}
		}
	}

	return nil, nil
}

func getRelativeItem(hist *historian.Historian, parentPID *int, offset int64) (*historian.Item, error) {
	filter := &historian.Filter{
		Reverse: offset < 0,
	}

	var count int64
	if filter.Reverse {
		offset = -offset
	} else {
		count = -1
	}

	var found *historian.Item

	err := hist.Each(filter, func(item *historian.Item) error {
		if parentPID != nil && *item.ParentPID != *parentPID {
			return nil
		}

		count++
		if count > offset {
			return errors.New("item not found")
		}

		if offset == count {
			found = item
			return historian.ErrStop
		}

		return nil
	})

	return found, err
}

func getItem(hist *historian.Historian, id uint64) (*historian.Item, error) {
	filter := &historian.Filter{
		Reverse: true, // hope that most likely case will be grabbing a recent one
	}

	var found *historian.Item

	err := hist.Each(filter, func(item *historian.Item) error {
		if item.ID == id {
			found = item
			return historian.ErrStop
		}

		return nil
	})

	if found == nil && err == nil {
		return nil, errors.New("item not found")
	}

	return found, err
}

func openHistorian(shell string, readOnly bool) *historian.Historian {
	os.Mkdir(historyDir, 0700)
	path := filepath.Join(historyDir, shell+".db")
	hist, err := historian.Open(path, readOnly)
	if err != nil {
		panic(err)
	}

	if os.Getenv("AARDY_INTEGRATED") != "true" {
		tools.LogStdError("\nWARNING: History integration is not enabled from this shell.\n\n"+
			"Use the following to enable it:\n"+
			"  eval `%s init`\n\n", getAppExe())
	}

	return hist
}

func syncThisDir(root string) error {
	appExe := getAppExe()
	replacer := strings.NewReplacer("{{ AppName }}", app.Name, "{{ AppExe }}", appExe)

	return fs.WalkDir(thisDir, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		localPath := tools.MyPathTo(path)

		if d.IsDir() {
			err = os.Mkdir(localPath, 0700)
			if err != nil && !os.IsExist(err) {
				return err
			}

			return nil
		}

		body, err := thisDir.ReadFile(path)
		if err != nil {
			return err
		}

		contents := replacer.Replace(string(body))

		localInfo, err := os.Stat(localPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		if localInfo != nil && localInfo.Size() == int64(len(contents)) {
			// for now, if the size doesn't change we'll assume it's up-to-date
			return nil
		}

		os.Remove(localPath)
		localF, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY, 0400)
		if err != nil {
			return err
		}
		defer localF.Close()

		_, err = localF.WriteString(contents)
		if err != nil {
			return err
		}

		return nil
	})
}

func getAppExe() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	return exe
}
