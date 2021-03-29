package config

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
	"gopkg.in/yaml.v2"
	"os"
)

var (
	defaultConfigFile = os.ExpandEnv("$HOME/.pruney/pruney.yaml")
)

type Entry struct {
	ServiceUrl string `yaml:"service_url"`
}

type Config struct {
	Entries      map[string]Entry `yaml:"entries"`
	DefaultEntry string           `yaml:"default"`
}

func (c *Config) Add(name, serviceUrl string) error {
	_, found := c.Entries[name]
	if found {
		return fmt.Errorf("entry with name = %s exists", name)
	}
	c.Entries[name] = Entry{ServiceUrl: serviceUrl}
	c.DefaultEntry = name
	return nil
}

func (c *Config) Del(name string) error {
	_, found := c.Entries[name]
	if !found {
		return fmt.Errorf("entry with name = %s not found", name)
	}
	delete(c.Entries, name)
	if c.DefaultEntry == name {
		c.DefaultEntry = ""
	}
	return nil
}

func (c *Config) SetDefault(name string) error {
	_, found := c.Entries[name]
	if !found {
		return fmt.Errorf("entry with name = %s not found", name)
	}
	c.DefaultEntry = name
	return nil
}

func (c *Config) SaveFile(filename string) error {
	fp, err := tools.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("OpenFile %s err = %v", filename, err)
	}
	defer tools.CloseFile(fp)
	enc := yaml.NewEncoder(fp)
	return enc.Encode(&c)
}

func New(filename string) (*Config, error) {
	fp, err := tools.OpenFile(filename, os.O_RDONLY|os.O_CREATE)
	if err != nil {
		return nil, err
	}
	defer tools.CloseFile(fp)

	size, err := filseSize(fp)
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return &Config{
			Entries:      map[string]Entry{},
			DefaultEntry: "",
		}, nil
	}

	dec := yaml.NewDecoder(fp)
	var c Config
	if err := dec.Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func filseSize(fp *os.File) (int64, error) {
	fi, err := fp.Stat()
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func Default() (*Config, error) {
	return New(defaultConfigFile)
}

func CmdConfig(argv []string, version string) error {
	usage := `
usage: pruney config list 
       pruney config add <name> --service-url=<url>
       pruney config delete <name>
       pruney config set-default <name>

Options:
  --service-url=<url>  Service URL endpoint.
  -h --help            Show this screen.

Examples:
  List all configurations for this client
  $ pruney config list 

  Add a new configuration with the name "foobar" and service url: https://foobar.aardvarklabs.com:9443
  $ pruney config add foobar --service-url=https://foobar.aardvarklabs.com:9443

  Delete an existing configuration with name "foobar"
  $ pruney config delete foobar

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
	if tools.OptsBool(opts, "add") == true {
		if err := cfg.Add(tools.OptsStr(opts, "<name>"),
			tools.OptsStr(opts, "--service-url")); err != nil {
			return err
		}
		doSave = true
	} else if tools.OptsBool(opts, "list") == true {
		fmt.Printf("Listing entries\n")
	} else if tools.OptsBool(opts, "delete") == true {
		if err := cfg.Del(tools.OptsStr(opts, "<name>")); err != nil {
			return err
		}
		doSave = true
	} else if tools.OptsBool(opts, "set-default") == true {
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
