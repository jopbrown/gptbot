package cfgs

import (
	"embed"
	"io"
	"os"
	"path/filepath"
	"time"

	"dario.cat/mergo"
	"github.com/jopbrown/gobase/errors"
	"github.com/jopbrown/gobase/fsutil"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DebugMode bool `yaml:"DebugMode"`

	ChatGptApiUrl           string `yaml:"ChatGptApiUrl"`
	ChatGptAccessToken      string `yaml:"ChatGptAccessToken"`
	ChatGptModel            string `yaml:"ChatGptModel"`
	ChatGptHideThinkSection bool   `yaml:"ChatGptHideThinkSection"`

	SessionExpirePeriod  time.Duration   `yaml:"SessionExpirePeriod"`
	SessionClearInterval time.Duration   `yaml:"SessionClearInterval"`
	Bots                 map[string]*Bot `yaml:"Bots"`
	Roles                Roles           `yaml:"Roles"`
	ServePort            int             `yaml:"ServePort"`
	MaxTaskQueueCap      int             `yaml:"MaxTaskQueueCap"`
	LogPath              string          `yaml:"LogPath"`
	CmdsTalkToAI         []string        `yaml:"CmdsTalkToAI"`
	CmdsClearSession     []string        `yaml:"CmdsClearSession"`
	CmdsChangeRole       []string        `yaml:"CmdsChangeRole"`
}

type Bot struct {
	DefaultRole       string `yaml:"DefaultRole"`
	LineChannelToken  string `yaml:"LineChannelToken"`
	LineChannelSecret string `yaml:"LineChannelSecret"`
}

//go:embed default
var defaultCfgFs embed.FS

func DefaultConfig() *Config {
	r := errors.Must1(defaultCfgFs.Open("default/config.yaml"))
	cfg := errors.Must1(ReadConfig(r))
	cfg.Roles = DefaultRoles()
	cfg.LogPath = filepath.Join(fsutil.AppDir(), "logs")
	return cfg
}

func LoadConfig(fname string) (*Config, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, errors.ErrorAt(err)
	}
	defer f.Close()

	cfg, err := ReadConfig(f)
	if err != nil {
		return nil, errors.ErrorAt(err)
	}

	return cfg, nil
}

func ReadConfig(r io.Reader) (*Config, error) {
	cfg := &Config{}
	err := yaml.NewDecoder(r).Decode(cfg)
	if err != nil {
		return nil, errors.ErrorAt(err)
	}

	return cfg, nil
}

func (cfg *Config) Merge(cfg2 *Config) error {
	err := mergo.Merge(cfg, cfg2)
	if err != nil {
		return errors.ErrorAt(err)
	}
	return nil
}

func (cfg *Config) MergeDefault() error {
	return cfg.Merge(DefaultConfig())
}

func (cfg *Config) SaveConfig(fname string) error {
	f, err := fsutil.OpenFileWrite(fname)
	if err != nil {
		return errors.ErrorAt(err)
	}
	defer f.Close()

	err = cfg.WriteConfig(f)
	if err != nil {
		return errors.ErrorAt(err)
	}

	return nil
}

func (cfg *Config) WriteConfig(w io.Writer) error {
	err := yaml.NewEncoder(w).Encode(cfg)
	if err != nil {
		return errors.ErrorAt(err)
	}

	return nil
}
