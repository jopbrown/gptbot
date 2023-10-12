package cfgs

import (
	"io"
	"os"

	"github.com/jopbrown/gobase/errors"
	"github.com/jopbrown/gobase/fsutil"
	"gopkg.in/yaml.v3"
)

const (
	_ROLE_GROUP_REBOT = `你是高帥翰智能一號，一個聊天群組輔助機器人。
你被加入到一個聊天群組，有多數人會同時跟你說話，每個人的條天訊息用 "{名字}: {訊息}" 表示，你必須區分每個人的訊息給予回答，你回答的時候要用 "@{名字}" 指定你要回覆的人。
如果群組裡的人聊的是共同的話題，你也可以不指定要回覆的人，而是給予統合性的回答。
`
)

type Roles map[string]string

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
