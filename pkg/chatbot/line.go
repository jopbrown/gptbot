package chatbot

import (
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/jopbrown/gobase/errors"
	"github.com/jopbrown/gobase/log"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func (bot *Bot) linebotCallback(c *gin.Context) {
	events, err := bot.lineClient.ParseRequest(c.Request)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			c.JSON(http.StatusBadRequest, gin.H{"message": linebot.ErrInvalidSignature})
			return
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				msg := strings.TrimLeftFunc(message.Text, unicode.IsSpace)
				if _, ok := messageMatchCmd(msg, bot.cfg.CmdsClearSession); ok {
					bot.taskQueue <- &ClearSessionTask{
						SessionID: lineGetSessionID(event),
						ReplyFn:   bot.lineReplyFnWithToken(event.ReplyToken),
					}
				} else if role, ok := messageMatchCmd(msg, bot.cfg.CmdsChangeRole); ok {
					bot.taskQueue <- &ChangeRoleTask{
						SessionID: lineGetSessionID(event),
						Role:      role,
						ReplyFn:   bot.lineReplyFnWithToken(event.ReplyToken),
					}
				} else {
					trimedMsg, isTalkToAI := messageMatchCmd(msg, bot.cfg.CmdsTalkToAI)
					if isTalkToAI || !lineIsGroupEvent(event) {
						userName, err := bot.lineGetUserName(event.Source.UserID)
						if err != nil {
							log.ErrorAt(err)
							continue
						}
						log.Debug("userName: ", userName)

						bot.taskQueue <- &ChatTask{
							UserName:  userName,
							SessionID: lineGetSessionID(event),
							Message:   trimedMsg,
							ReplyFn:   bot.lineReplyFnWithToken(event.ReplyToken),
						}
					}
				}

			default:
				// ignore other message types
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (bot *Bot) lineReplyFnWithToken(token string) func(string, ...string) error {
	return func(reply string, imgUrls ...string) error {
		msgs := make([]linebot.SendingMessage, 0, 1+len(imgUrls))
		msgs = append(msgs, linebot.NewTextMessage(reply))
		for _, url := range imgUrls {
			msgs = append(msgs, linebot.NewImageMessage(url, url))
		}
		if _, err := bot.lineClient.ReplyMessage(token, msgs...).Do(); err != nil {
			return errors.ErrorAt(err)
		}
		return nil
	}
}

func (bot *Bot) lineGetUserName(userID string) (string, error) {
	if userName, ok := bot.userNameCache[userID]; ok {
		return userName, nil
	}
	profile, err := bot.lineClient.GetProfile(userID).Do()
	if err != nil {
		log.Warnf(errors.GetErrorDetails(errors.ErrorAtf(err, "unable to get user profile: %s", userID)))
		return "路人甲", nil
	}
	userName := profile.DisplayName
	bot.userNameCache[userID] = userName

	return userName, nil
}

func lineIsGroupEvent(event *linebot.Event) bool {
	switch event.Source.Type {
	case linebot.EventSourceTypeRoom, linebot.EventSourceTypeGroup:
		return true
	}

	return false
}

func lineGetSessionID(event *linebot.Event) string {
	switch event.Source.Type {
	case linebot.EventSourceTypeRoom:
		return event.Source.RoomID
	case linebot.EventSourceTypeGroup:
		return event.Source.GroupID
	}

	return event.Source.UserID
}
