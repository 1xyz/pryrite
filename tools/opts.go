package tools

import (
	"github.com/docopt/docopt-go"
	"time"
)

func OptsBool(opts docopt.Opts, key string) bool {
	v, err := opts.Bool(key)
	if err != nil {
		Log.Fatal().Msgf("OptsBool: %v parse err = %v", key, err)
	}
	return v
}

func OptsStr(opts docopt.Opts, key string) string {
	v, err := opts.String(key)
	if err != nil {
		Log.Fatal().Msgf("OptsStr: %v parse err = %v", key, err)
	}
	return v
}

func OptsInt(opts docopt.Opts, key string) int {
	v, err := opts.Int(key)
	if err != nil {
		Log.Fatal().Msgf("OptsInt: %v parse err = %v", key, err)
	}
	return v
}

func OptsSeconds(opts docopt.Opts, key string) time.Duration {
	v := OptsInt(opts, key)
	return time.Duration(v) * time.Second
}
