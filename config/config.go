package config

import (
	"fmt"
	"os"
	"time"

	"github.com/1xyz/pryrite/tools"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

var (
	DefaultConfigFile = tools.MyPathTo("pryrite.yaml")
)

const (
	DefaultDashboardURL = "https://foo/bar"
	DefaultServiceURL   = "https://foo/bar"
)

type Entry struct {
	Name             string                   `yaml:"name"`
	Mode             string                   `yaml:"mode"`
	DashboardUrl     string                   `yaml:"dashboard_url"`
	ServiceUrl       string                   `yaml:"service_url"`
	LastUpdateCheck  time.Time                `yaml:"last_update_check"`
	AuthScheme       string                   `yaml:"auth_scheme"`
	Email            string                   `yaml:"email"`
	ClientID         string                   `yaml:"client_id"`
	SkipSSLCheck     bool                     `yaml:"skip_ssl_check"`
	Style            string                   `yaml:"style"`
	ExecutionTimeout tools.MarshalledDuration `yaml:"execution_timeout"`
	HideInspectIntro bool                     `yaml:"hide_inspect_intro"`
}

type Config struct {
	Entries      []Entry `yaml:"entries"`
	DefaultEntry string  `yaml:"default"`
}

func (c *Config) Get(name string) (*Entry, bool) {
	index, found := c.getIndex(name)
	if !found {
		return nil, found
	}

	return &c.Entries[index], found
}

func (c *Config) GetDefaultEntry() (*Entry, bool) {
	return c.Get(c.DefaultEntry)
}

func (c *Config) Set(e *Entry) error {
	index, found := c.getIndex(e.Name)
	if !found {
		return fmt.Errorf("entry not found")
	}
	c.Entries[index] = *e
	return nil
}

func (c *Config) Add(name, serviceUrl string) error {
	_, found := c.getIndex(name)
	if found {
		return fmt.Errorf("entry with name = %s exists", name)
	}
	c.Entries = append(c.Entries,
		Entry{
			Name:       name,
			ServiceUrl: serviceUrl,
			ClientID:   uuid.New().String(),
		},
	)
	c.DefaultEntry = name
	return nil
}

func (c *Config) Del(name string) error {
	index, found := c.getIndex(name)
	if !found {
		return fmt.Errorf("entry with name = %s not found", name)
	}

	c.Entries = append(c.Entries[:index], c.Entries[index+1:]...)
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

func (c *Config) getIndex(name string) (int, bool) {
	for i, e := range c.Entries {
		if e.Name == name {
			return i, true
		}
	}
	return -1, false
}

func New(filename string) (*Config, error) {
	fp, err := tools.OpenFile(filename, os.O_RDONLY|os.O_CREATE)
	if err != nil {
		return nil, err
	}
	defer tools.CloseFile(fp)

	size, err := fileSize(fp)
	if err != nil {
		return nil, err
	}

	if size == 0 {
		e := Entry{
			Name:             "Default",
			LastUpdateCheck:  time.Now(),
			Email:            "unknown@bar.com",
			ExecutionTimeout: tools.MarshalledDuration{Duration: 0 * time.Second},
			HideInspectIntro: false,
		}
		c := &Config{
			Entries:      []Entry{e},
			DefaultEntry: e.Name,
		}
		if err := c.Set(&e); err != nil {
			return nil, err
		}
		return c, nil
	}

	dec := yaml.NewDecoder(fp)
	var c Config
	if err := dec.Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func fileSize(fp *os.File) (int64, error) {
	fi, err := fp.Stat()
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func Default() (*Config, error) {
	return New(DefaultConfigFile)
}

func GetEntry(name string) (*Entry, error) {
	cfg, err := Default()
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = cfg.DefaultEntry
		if name == "" {
			return nil, fmt.Errorf("there is no default configuration. See  Config --help to set a default configuration")
		}
	}

	entry, found := cfg.Get(name)
	if !found {
		return nil, fmt.Errorf("configuration with name = %v not found", name)
	}
	return entry, nil
}

func SetEntry(e *Entry) error {
	cfg, err := Default()
	if err != nil {
		return err
	}
	if err := cfg.Set(e); err != nil {
		return err
	}
	return cfg.SaveFile(DefaultConfigFile)
}

func CreateDefaultConfigIfEmpty() error {
	cfg, err := Default()
	if err != nil {
		return err
	}
	if len(cfg.Entries) > 0 {
		return nil
	}
	if err := cfg.Add("remote", DefaultServiceURL); err != nil {
		return err
	}
	return cfg.SaveFile(DefaultConfigFile)
}
