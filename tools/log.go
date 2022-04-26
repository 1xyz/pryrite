package tools

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

var (
	Log           = zlog.Logger
	logConfigFile = MyPathTo("logging.yaml")
	logFile       = MyPathTo("activity.log")
	traceLabels   = map[string]string{}
)

type rollingLogConfig struct {
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
	Filename   string `yaml:"filename"`
}

// createLogConfigFile creates a default rolling log config file
// if ones does not exists
func createLogConfigFile() error {
	exists, err := StatExists(logConfigFile)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	c := &rollingLogConfig{
		MaxSizeMB:  5,
		MaxBackups: 2,
		MaxAgeDays: 30,
		Filename:   logFile,
	}

	fp, err := OpenFile(logConfigFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("OpenFile %s err = %v", logConfigFile, err)
	}
	defer CloseFile(fp)
	enc := yaml.NewEncoder(fp)
	return enc.Encode(&c)
}

// loadRollingLogConfig loads a rolling log configuration from file
func loadRollingLogConfig() (*rollingLogConfig, error) {
	if err := createLogConfigFile(); err != nil {
		return nil, err
	}

	fp, err := OpenFile(logConfigFile, os.O_RDONLY)
	if err != nil {
		return nil, err
	}

	dec := yaml.NewDecoder(fp)
	c := rollingLogConfig{}
	if err := dec.Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

// OpenLogger opens the default rolling file logger
// and sets the verbosity of the logger
func OpenLogger(verbose bool) (io.Closer, error) {
	// create a log config if needed
	if err := createLogConfigFile(); err != nil {
		return nil, err
	}
	// load the config from file
	c, err := loadRollingLogConfig()
	if err != nil {
		return nil, err
	}
	// ensure that the log directory is present
	if err := EnsureDir(filepath.Dir(c.Filename)); err != nil {
		return nil, err
	}
	// create log writer
	w := &lumberjack.Logger{
		Filename:   c.Filename,
		MaxBackups: c.MaxBackups,
		MaxSize:    c.MaxSizeMB,
		MaxAge:     c.MaxAgeDays,
	}
	// set log level
	level := zerolog.InfoLevel
	if verbose {
		level = zerolog.DebugLevel
	}
	// initialize logger
	Log = zlog.Output(w).Level(level)
	// set standard logger up to use zlog (for 3rd parties that use it, like selfupdate)
	log.SetFlags(0) // remove timestamps, etc., since zlog handles that for us
	log.SetOutput(&dumbLogWriter{})
	labelsStr := strings.ToLower(strings.TrimSpace(os.Getenv("AARDY_TRACE")))
	if labelsStr != "" {
		labels := strings.Split(labelsStr, ",")
		for _, label := range labels {
			traceLabels[label] = fmt.Sprintf("TRACE(%s): ", label)
		}
		if len(labels) > 0 {
			Trace = traceLog
			Log.Debug().Str("labels", labelsStr).Msg("TRACE logs enabled")
		}
	}
	return w, nil
}

func TrimLength(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func RemoveNewLines(s string, r string) string {
	s = strings.ReplaceAll(s, "\n", r)
	s = strings.ReplaceAll(s, "\r", r)
	return s
}

func FmtTime(t *time.Time) string {
	return t.In(time.Local).Format("Jan _2 3:04PM")
}

func LogStdout(format string, v ...interface{}) {
	_, err := fmt.Fprintf(os.Stdout, format, v...)
	if err != nil {
		panic(err)
	}
	Log.Info().Msgf(format, v...)
}

func LogStdError(format string, v ...interface{}) {
	_, err := fmt.Fprintf(os.Stderr, format, v...)
	if err != nil {
		panic(err)
	}
	Log.Error().Msgf(format, v...)
}

func LogStderr(err error, format string, v ...interface{}) {
	Log.Err(err).Msgf(format, v...)
	_, fErr := fmt.Fprintf(os.Stderr, format, v...)
	if fErr != nil {
		panic(fErr)
	}
}

func LogStderrExit(err error, format string, v ...interface{}) {
	LogStderr(err, format, v...)
	os.Exit(1)
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	Log.Printf("%s took %s", name, elapsed)
}

var Trace = func(label string, msg string, vals ...interface{}) {}

func traceLog(label, msg string, vals ...interface{}) {
	lbl, ok := traceLabels[label]
	if !ok {
		return
	}
	fmt := fmt.Sprint(lbl, msg, ":", strings.Repeat(" %v", len(vals)))
	Log.Printf(fmt, vals...)
}

//--------------------------------------------------------------------------------

type dumbLogWriter struct{}

func (dlw *dumbLogWriter) Write(data []byte) (int, error) {
	msg := strings.TrimRight(string(data), "\r\n")
	Log.Info().Msg(msg)
	return len(data), nil
}
