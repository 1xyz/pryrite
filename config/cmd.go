package config

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
)

func Cmd(argv []string, version string) error {
	usage := `The "config" command provides options to configure pruney 

usage: pruney config add <name> --service-url=<url>
       pruney config list
       pruney config remove <name>
       pruney config set-default <name>

Options:
  --service-url=<url>  Service URL endpoint.
  -h --help            Show this screen.

Examples:
  Add a new configuration with the name "foobar" and service url: https://foobar.aardvarklabs.com:9443
  $ pruney config add foobar --service-url=https://foobar.aardvarklabs.com:9443

  List all configurations for this client
  $ pruney config list

  Remove an existing configuration with name "foobar"
  $ pruney config remove foobar

  Set the default configuration to "foobar"
  $ pruney config set-default foobar
`
	opts, err := docopt.ParseArgs(usage, argv, version)
	if err != nil {
		tools.Log.Fatal().Msgf("error parsing arguments. err=%v", err)
	}

	cfg, err := Default()
	if err != nil {
		return err
	}

	tools.Log.Info().Msgf("Config: %v add = %v", cfg, tools.OptsBool(opts, "add"))
	doSave := false
	if tools.OptsBool(opts, "add") {
		if err := cfg.Add(tools.OptsStr(opts, "<name>"),
			tools.OptsStr(opts, "--service-url")); err != nil {
			return err
		}
		doSave = true
	} else if tools.OptsBool(opts, "list") {
		fmt.Printf("Listing entries\n")
	} else if tools.OptsBool(opts, "remove") {
		if err := cfg.Del(tools.OptsStr(opts, "<name>")); err != nil {
			return err
		}
		doSave = true
	} else if tools.OptsBool(opts, "set-default") {
		if err := cfg.SetDefault(tools.OptsStr(opts, "<name>")); err != nil {
			return err
		}
		doSave = true
	}

	if doSave {
		if err := cfg.SaveFile(defaultConfigFile); err != nil {
			return err
		}
	}

	tr := &tableRender{config: cfg}
	tr.Render()
	return nil
}
