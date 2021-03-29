package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

//  os.O_RDWR|os.O_CREATE
func OpenFile(filename string, flag int) (*os.File, error) {
	if err := EnsureDir(filepath.Dir(filename)); err != nil {
		return nil, err
	}
	fp, err := os.OpenFile(filename, flag, 0644)
	if err != nil {
		return nil, fmt.Errorf("os.open %s err = %v", filename, err)
	}
	return fp, nil
}

func EnsureDir(dirname string) error {
	if err := os.MkdirAll(dirname, 0755); err != nil {
		return fmt.Errorf("os.MkdirAll %s err = %v", dirname, err)
	}
	return nil
}

func CloseFile(fp *os.File) {
	if err := fp.Close(); err != nil {
		Log.Warn().Msgf("close file %s  err = %v", fp.Name(), err)
	}
}
