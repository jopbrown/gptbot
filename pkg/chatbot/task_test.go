package chatbot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getImageUrlsFromReply(t *testing.T) {
	reply, urls := getImageUrlsFromReply(`@JB 這裡是您要的圖片，https://image.pollinations.ai/prompt/drunken-man-dancing。希望您會喜歡。`)
	assert.Equal(t, "@JB 這裡是您要的圖片，圖1。希望您會喜歡。", reply)
	assert.Equal(t, []string{"https://image.pollinations.ai/prompt/drunken-man-dancing"}, urls)
}
