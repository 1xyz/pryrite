package log

import (
	"encoding/json"
	"fmt"
	"github.com/aardlabs/terminal-poc/tools"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
)

// log on the file system
type fsLog struct {
	dir string
}

func (l *fsLog) Len() (int, error) {
	files, err := l.getLogFiles()
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

func (l *fsLog) Append(entry *ResultLogEntry) error {
	filename := getfilename(entry.ID)
	fileWithPath := path.Join(l.dir, filename)

	// Create one file per log. if an entry exists, then overwrite it
	// a logID is expected be  unique within the context of a node
	f, err := tools.OpenFile(fileWithPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return err
	}
	defer tools.CloseFile(f)

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = f.Write(jsonBytes)
	return err
}

func (l *fsLog) Each(cb func(int, *ResultLogEntry) bool) error {
	files, err := l.getLogFiles()
	if err != nil {
		return err
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})
	for i := range files {
		tools.Log.Info().Msgf("Each: %d entry %s", i, files[i].Name())
		entry, err := l.decode(files[i].Name())
		if err != nil {
			return err
		}
		if !cb(i, entry) {
			return nil
		}
	}
	return nil
}

// Find scans the slice of BlockExecutionResults by ID
func (l *fsLog) Find(id string) (*ResultLogEntry, error) {
	files, err := l.getLogFiles()
	if err != nil {
		return nil, err
	}

	filename := getfilename(id)
	for _, f := range files {
		if f.Name() == filename {
			entry, err := l.decode(f.Name())
			if err != nil {
				return nil, err
			}
			return entry, nil
		}
	}

	return nil, ErrResultLogEntryNotFound
}

func (l *fsLog) decode(filename string) (*ResultLogEntry, error) {
	fileWithPath := path.Join(l.dir, filename)
	fr, err := tools.OpenFile(fileWithPath, os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer tools.CloseFile(fr)

	var r ResultLogEntry
	if err := json.NewDecoder(fr).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (l *fsLog) getLogFiles() ([]fs.FileInfo, error) {
	files, err := ioutil.ReadDir(l.dir)
	if err != nil {
		return nil, err
	}

	var logfiles []fs.FileInfo
	for i := range files {
		f := files[i]
		if f.IsDir() || !strings.HasPrefix(f.Name(), "log_") || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		logfiles = append(logfiles, files[i])
	}

	return files, nil
}

func getfilename(logID string) string {
	return fmt.Sprintf("log_%s.json", logID)
}

type fsLogIndex struct {
	dir string
}

func newFSLogIndex(logDir string) (*fsLogIndex, error) {
	if err := tools.EnsureDir(logDir); err != nil {
		return nil, err
	}
	return &fsLogIndex{
		dir: logDir,
	}, nil
}

func (i *fsLogIndex) Append(entry *ResultLogEntry) error {
	fsLog, err := i.getOrCreateLog(entry.NodeID)
	if err != nil {
		return err
	}
	return fsLog.Append(entry)
}

func (i *fsLogIndex) Get(nodeID string) (ResultLog, error) {
	dir := path.Join(i.dir, nodeID)
	if exists, err := tools.StatExists(dir); err != nil {
		return nil, err
	} else if !exists {
		return nil, ErrResultLogNotFound
	}
	return &fsLog{dir: dir}, nil
}

func (i *fsLogIndex) getOrCreateLog(nodeID string) (*fsLog, error) {
	dir := path.Join(i.dir, nodeID)
	if err := tools.EnsureDir(dir); err != nil {
		return nil, err
	}
	return &fsLog{dir: dir}, nil
}
