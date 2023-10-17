package chatbot

import (
	"net/http"

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
