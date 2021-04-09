package log

import "github.com/atotto/clipboard"

func clipTo(text string) error {
	return clipboard.WriteAll(text)
}

func getClip() (string, error) {
	return clipboard.ReadAll()
}
