package capture

import (
	"github.com/aardlabs/terminal-poc/tools"
	"os"
	"time"
)

func Play(filename string) error {
	fSet, err := ReadFromFile(filename)
	if err != nil {
		return err
	}
	tools.Log.Info().Msgf("Play: number of entries =  %v", len(fSet.Stdout))
	if err := play(fSet); err != nil {
		return err
	}
	return nil
}

func play(fSet *FrameSet) error {
	for _, entry := range fSet.Stdout {
		time.Sleep(time.Duration(float64(time.Second) * entry.Delay))
		if err := write(entry.Data); err != nil {
			//log.Warnf("error = %v", err)
		}
	}
	return nil
}

func write(data []byte) error {
	_, err := os.Stdout.Write(data)
	if err != nil {
		return err
	}
	err = os.Stdout.Sync()
	if err != nil {
		return err
	}
	return nil
}
