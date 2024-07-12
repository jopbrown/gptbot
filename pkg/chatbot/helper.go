package chatbot

import (
	"net/http"
	"strings"

	"github.com/jopbrown/gobase/errors"
	"github.com/sashabaranov/go-openai"
)

func GetOpenAIErrCode(err error) int {
	if reqErr, ok := errors.AsIs[*openai.RequestError](err); ok {
		return reqErr.HTTPStatusCode
	}

	if errRes, ok := errors.AsIs[*openai.APIError](err); ok {
		return errRes.HTTPStatusCode
	}

	return http.StatusOK
}

func messageMatchCmd(msg string, cmds []string) (string, bool) {
	if len(cmds) == 0 {
		return msg, true
	}

	for _, cmd := range cmds {
		// empty command means not need to prefix anything
		if cmd == "" {
			return msg, true
		}

		if cmd[0] == '/' && msg[1] != '/' {
			cmd = cmd[1:]
		}

		if len(msg) < len(cmd) {
			continue
		}

		if strings.EqualFold(msg[:len(cmd)], cmd) {
			if len(msg) == len(cmd) {
				return msg, true
			}
			return strings.TrimSpace(msg[len(cmd):]), true
		}
	}
	return msg, false
}
