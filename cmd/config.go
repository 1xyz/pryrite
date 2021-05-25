package cmd

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
)

func ConfigCmd(params *Params) error {
	usage := `The "config" command provides options to configure aard 

usage: aard config add <name> --service-url=<url>
       aard config list
       aard config remove <name>
       aard config set-default <name>

Options:
  --service-url=<url>  Service URL endpoint.
  -h --help            Show this screen.

Examples:
  Register a new configuration with the name "foobar" and service url: https://foobar.aardvarklabs.com:9443
  $ aard config add foobar --service-url=https://foobar.aardvarklabs.com:9443

  List all configurations for this client
  $ aard config list

  Remove an existing configuration with name "foobar"
  $ aard config remove foobar

  Set the default configuration to "foobar"
  $ aard config set-default foobar
`
	opts, err := docopt.ParseArgs(usage, params.Argv, params.Version)
	if err != nil {
		tools.Log.Fatal().Msgf("error parsing arguments. err=%v", err)
	}

	cfg, err := config.Default()
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
		if err := cfg.SaveFile(config.DefaultConfigFile); err != nil {
			return err
		}
	}

	tr := &config.TableRender{Config: cfg}
	tr.Render()
	return nil
}
