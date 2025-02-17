package chatbot

import (
	"net/http"
	"path"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/jopbrown/gobase/errors"
	"github.com/jopbrown/gobase/log"
	"github.com/line/line-bot-sdk-go/v8/linebot"
)

func (bot *Bot) linebotCallback(c *gin.Context) {
	fpath := c.FullPath()
	client := bot.lineClients[fpath]

	defaultRole := bot.cfg.Bots[c.FullPath()].DefaultRole

	events, err := client.ParseRequest(c.Request)
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
			sessionID := path.Join(fpath, lineGetSessionID(event))
			session := bot.sessMgr.GetSession(sessionID, defaultRole)
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				msg := strings.TrimLeftFunc(message.Text, unicode.IsSpace)
				if _, ok := messageMatchCmd(msg, bot.cfg.CmdsClearSession); ok {
					bot.taskQueue <- &ClearSessionTask{
						Session: session,
						ReplyFn: bot.lineReplyFnWithToken(client, event.ReplyToken),
					}
				} else if role, ok := messageMatchCmd(msg, bot.cfg.CmdsChangeRole); ok {
					bot.taskQueue <- &ChangeRoleTask{
						Session: session,
						Role:    role,
						ReplyFn: bot.lineReplyFnWithToken(client, event.ReplyToken),
					}
				} else {
					userName, err := bot.lineGetUserName(client, event.Source.UserID)
					if err != nil {
						log.ErrorAt(err)
						continue
					}
					botName, err := bot.lineGetBotName(client, fpath)
					if err != nil {
						log.ErrorAt(err)
						continue
					}
					bot.taskQueue <- &ChatTask{
						UserName: userName,
						BotName:  botName,
						Session:  session,
						Message:  msg,
						IsGroup:  lineIsGroupEvent(event),
						ReplyFn:  bot.lineReplyFnWithToken(client, event.ReplyToken),
					}
				}

			default:
				// ignore other message types
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (bot *Bot) lineReplyFnWithToken(client *linebot.Client, token string) func(string, ...string) error {
	return func(reply string, imgUrls ...string) error {
		msgs := make([]linebot.SendingMessage, 0, 1+len(imgUrls))
		msgs = append(msgs, linebot.NewTextMessage(reply))
		for _, url := range imgUrls {
			msgs = append(msgs, linebot.NewImageMessage(url, url))
		}
		if _, err := client.ReplyMessage(token, msgs...).Do(); err != nil {
			return errors.ErrorAt(err)
		}
		return nil
	}
}

func (bot *Bot) lineGetBotName(client *linebot.Client, fpath string) (string, error) {
	if userName, ok := bot.userNameCache[fpath]; ok {
		return userName, nil
	}
	botInfo, err := client.GetBotInfo().Do()
	if err != nil {
		log.Warn(errors.GetErrorDetails(errors.ErrorAtf(err, "unable to get bot info: %q", fpath)))
	}
	userName := botInfo.DisplayName
	bot.userNameCache[fpath] = userName

	return userName, nil
}

func (bot *Bot) lineGetUserName(client *linebot.Client, userID string) (string, error) {
	if userName, ok := bot.userNameCache[userID]; ok {
		return userName, nil
	}
	profile, err := client.GetProfile(userID).Do()
	if err != nil {
		log.Warn(errors.GetErrorDetails(errors.ErrorAtf(err, "unable to get user profile: %s", userID)))
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
