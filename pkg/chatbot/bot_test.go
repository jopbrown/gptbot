package chatbot

import (
	"testing"

	"github.com/jopbrown/gobase/errors"
	"github.com/liuzl/gocc"
	"github.com/stretchr/testify/assert"
)

func Test_messageMatchCmd(t *testing.T) {
	_, match := messageMatchCmd("清空", []string{"/清空"})
	assert.True(t, match)

	role, _ := messageMatchCmd("扮演群組聊天機器人", []string{"/扮演"})
	role = errors.Must1(errors.Must1(gocc.New("s2tw")).Convert(role))
	assert.Equal(t, "群組聊天機器人", role)
}
