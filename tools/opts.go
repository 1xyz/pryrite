package tools

import (
	"github.com/docopt/docopt-go"
	"github.com/rs/zerolog/log"
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

func OptsStrSlice(opts docopt.Opts, key string) []string {
	v, ok := opts[key]
	if !ok {
		log.Fatal().Msgf("OptsStrSlice: cannot find entry w/ key %v", key)
	}
	ss, ok := v.([]string)
	if !ok {
		log.Fatal().Msgf("OptsStrSlice: cannot cast %v to []string", v)
	}
	return ss
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

func OptsContains(opts docopt.Opts, key string) bool {
	entry, found := opts[key]
	return found && entry != nil
}
