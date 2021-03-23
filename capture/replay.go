package capture

import (
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func Replay(filename string) error {
	fSet, err := ReadFromFile(filename)
	if err != nil {
		return err
	}
	log.Infof("Replay: number of entries =  %v", len(fSet.Stdout))
	if err := play(fSet); err != nil {
		return err
	}
	return nil
}

func play(fSet *FrameSet) error {
	prevDelay := 0.0
	for _, entry := range fSet.Stdout {
		delay := entry.Delay - prevDelay
		time.Sleep(time.Duration(float64(time.Second) * delay))
		if err := write(entry.Data); err != nil {
			//log.Warnf("error = %v", err)
		}
		prevDelay = entry.Delay
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
