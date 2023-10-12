package cfgs

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	DefaultConfig().SaveConfig("tmp/config.yaml")
}
