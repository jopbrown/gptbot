package cfgs

import (
	"io"
	"os"

	"github.com/jopbrown/gobase/errors"
	"github.com/jopbrown/gobase/fsutil"
	"gopkg.in/yaml.v3"
)

type Role struct {
	Prompt               string   `yaml:"Prompt"`
	MaxConversationCount int      `yaml:"MaxConversationCount"`
	PrefixUserName       bool     `yaml:"PrefixUserName"`
	NotNeedSlashCmd      bool     `yaml:"NotNeedSlashCmd"`
	CmdsTalkToAI         []string `yaml:"CmdsTalkToAI"`
}

type Roles map[string]*Role

func DefaultRoles() Roles {
	r := errors.Must1(defaultCfgFs.Open("default/roles.yaml"))
	roles := errors.Must1(ReadRoles(r))
	return roles
}

func LoadRoles(fname string) (Roles, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, errors.ErrorAt(err)
	}
	defer f.Close()

	roles, err := ReadRoles(f)
	if err != nil {
		return nil, errors.ErrorAt(err)
	}

	return roles, nil
}

func ReadRoles(r io.Reader) (Roles, error) {
	roles := Roles{}
	err := yaml.NewDecoder(r).Decode(&roles)
	if err != nil {
		return nil, errors.ErrorAt(err)
	}
	return roles, nil
}

func (roles Roles) SaveRoles(fname string) error {
	f, err := fsutil.OpenFileWrite(fname)
	if err != nil {
		return errors.ErrorAt(err)
	}
	defer f.Close()

	err = roles.WriteRoles(f)
	if err != nil {
		return errors.ErrorAt(err)
	}

	return nil
}

func (roles Roles) WriteRoles(w io.Writer) error {
	err := yaml.NewEncoder(w).Encode(roles)
	if err != nil {
		return errors.ErrorAt(err)
	}

	return nil
}
