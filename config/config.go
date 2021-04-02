package config

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/tools"
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

func (c *Config) Get(name string) (*Entry, bool) {
	e, ok := c.Entries[name]
	return &e, ok
}

func (c *Config) Add(name, serviceUrl string) error {
	_, found := c.Get(name)
	if found {
		return fmt.Errorf("entry with name = %s exists", name)
	}
	c.Entries[name] = Entry{ServiceUrl: serviceUrl}
	c.DefaultEntry = name
	return nil
}

func (c *Config) Del(name string) error {
	_, found := c.Get(name)
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
	_, found := c.Get(name)
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

func GetEntry(name string) (*Entry, error) {
	cfg, err := Default()
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = cfg.DefaultEntry
		if name == "" {
			return nil, fmt.Errorf("there is no default configuration. See pruney config --help to set a default configuration")
		}
	}

	entry, found := cfg.Get(name)
	if !found {
		return nil, fmt.Errorf("configuration with name = %v not found", name)
	}
	return entry, nil
}
