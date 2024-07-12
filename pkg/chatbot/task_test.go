package chatbot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getImageUrlsFromReply(t *testing.T) {
	reply, urls := getImageUrlsFromReply(`@user 這裡是您要的圖片，https://image.pollinations.ai/prompt/drunken-man-dancing。希望您會喜歡。`)
	assert.Equal(t, "@user 這裡是您要的圖片，圖1。希望您會喜歡。", reply)
	assert.Equal(t, []string{"https://image.pollinations.ai/prompt/drunken-man-dancing"}, urls)

	reply, urls = getImageUrlsFromReply(`@user:  這裡是您要的圖片 (https://image.pollinations.ai/prompt/a photo of a cute puppy)  這是您喜歡的可愛小狗嗎？🐶`)
	assert.Equal(t, "@user:  這裡是您要的圖片 (圖1)  這是您喜歡的可愛小狗嗎？🐶", reply)
	assert.Equal(t, []string{"https://image.pollinations.ai/prompt/a-photo-of-a-cute-puppy"}, urls)
}
