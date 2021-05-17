package tools

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

var (
	Log           = zlog.Logger
	logConfigFile = os.ExpandEnv("$HOME/.aardvark/logging.yaml")
	logFile       = os.ExpandEnv("$HOME/.aardvark/aard.log")
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
	return w, nil
}

func TrimLength(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func FmtTime(t *time.Time) string {
	return t.In(time.Local).Format("Jan _2 3:04PM")
}

func LogStderrExit(message string, err error) {
	Log.Err(err).Msgf("%s. error: %v.", message, err)
	_, fErr := fmt.Fprintf(os.Stderr, "%s. error: %v.\n", message, err)
	if fErr != nil {
		panic(fErr)
	}
	os.Exit(1)
}

func LogStdout(message string) {
	Log.Info().Msg(message)
	_, fErr := fmt.Fprintln(os.Stdout, message)
	if fErr != nil {
		panic(fErr)
	}
}
