package capture

import (
	"encoding/json"
	"fmt"
	"github.com/aardlabs/terminal-poc/tools"
	"os"
	"time"
)

// Frameset is a bunch of frames with metadata
type FrameSet struct {
	SessionID string        `json:"session_id"`
	Stdout    []*FrameEntry `json:"stdout"`
}

type FrameEntry struct {
	Delay float64
	Data  []byte
}

func (f *FrameEntry) MarshalJSON() ([]byte, error) {
	s, err := json.Marshal(string(f.Data))
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf(`[%.6f, %s]`, f.Delay, s)), nil
}

func (f *FrameEntry) UnmarshalJSON(data []byte) error {
	var x interface{}
	err := json.Unmarshal(data, &x)
	if err != nil {
		return err
	}

	f.Delay = x.([]interface{})[0].(float64)
	s := []byte(x.([]interface{})[1].(string))
	b := make([]byte, len(s))
	copy(b, s)
	f.Data = b
	return nil
}

type FrameSetWriter struct {
	outFile   string
	startTime time.Time
	frameSet  *FrameSet
}

func NewFrameSetWriter(sessionID, filename string) (*FrameSetWriter, error) {
	return &FrameSetWriter{
		startTime: time.Now(),
		frameSet: &FrameSet{
			SessionID: sessionID,
			Stdout:    []*FrameEntry{},
		},
		outFile: filename,
	}, nil
}

func (fw *FrameSetWriter) Write(p []byte) (int, error) {
	delay := time.Since(fw.startTime).Seconds()
	data := make([]byte, len(p))
	n := copy(data, p)
	fw.frameSet.Stdout = append(fw.frameSet.Stdout, &FrameEntry{
		Delay: delay,
		Data:  data,
	})
	return n, nil
}

func (fw *FrameSetWriter) Close() error {
	ofile, err := os.Create(fw.outFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := ofile.Close(); err != nil {
			tools.Log.Warn().Msgf("frameSetWriter.Close: err = %v", err)
		}
	}()

	enc := json.NewEncoder(ofile)
	return enc.Encode(fw.frameSet)
}

func ReadFromFile(filename string) (*FrameSet, error) {
	inFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := inFile.Close(); err != nil {
			tools.Log.Warn().Msgf("readFromFile infile.close(%s) err = %v", filename, err)
		}
	}()

	dec := json.NewDecoder(inFile)
	fSet := FrameSet{
		Stdout: []*FrameEntry{},
	}
	if err := dec.Decode(&fSet); err != nil {
		return nil, fmt.Errorf("dec.Decode err = %v", err)
	}
	return &fSet, nil
}
