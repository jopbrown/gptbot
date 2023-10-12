package chatbot

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jopbrown/gobase/errors"
	"github.com/jopbrown/gobase/log"
	"github.com/jopbrown/gobase/log/rotate"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type Task interface {
	Do(bot *Bot) error
}

type ChatTask struct {
	UserName  string
	SessionID string
	Message   string
	ReplyFn   func(reply string, imgUrls ...string) error
}

func (task *ChatTask) Do(bot *Bot) error {
	log.Debugf("do chat task...\n %+v", task)

	s := bot.sessMgr.GetSession(task.SessionID)
	recorder, err := rotate.OpenFile(filepath.Join(bot.cfg.LogPath, "chats", fmt.Sprintf("chat-%s.txt", s.ShortID())), 24*time.Hour, 0)
	if err != nil {
		return errors.ErrorAt(err)
	}
	defer recorder.Close()

	role := bot.cfg.Roles[s.Role]
	if len(s.Messages) == 0 && len(role) != 0 {
		log.Debug("append system message ...")
		s.AddMessage(&openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: role,
		})
		fmt.Fprintln(recorder, "#####################")
		fmt.Fprintln(recorder, role)
	}

	msg := task.Message
	toMsg := fmt.Sprintf("%s: %s", task.UserName, msg)
	if s.Role == "女僕" || s.Role == "聊天機器人" {
		msg = toMsg
	}
	log.Info(msg)
	fmt.Fprintln(recorder, toMsg)

	chatMsg := &openai.ChatCompletionMessage{}
	chatMsg.Content = msg
	chatMsg.Role = openai.ChatMessageRoleUser
	s.AddMessage(chatMsg)

	log.Debug("send message to chatgpt ...")
	resp, err := bot.gptClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: s.Messages,
		},
	)

	if err != nil {
		if task.ReplyFn != nil {
			err1 := task.ReplyFn(fmt.Sprintf("小愛壞掉了:\n%s\n如果問題持續發生，建議清空。", errors.GetErrorDetails(err)))
			if err1 != nil {
				return errors.ErrorAt(errors.Join(err, err1))
			}
		}
		return errors.ErrorAt(err)
	}

	respMsg := &openai.ChatCompletionMessage{}
	respMsg.Role = openai.ChatMessageRoleAssistant
	respMsg.Content = resp.Choices[0].Message.Content

	s.AddMessage(respMsg)
	log.Info("AI:", respMsg.Content)
	fmt.Fprintln(recorder, "AI:", respMsg.Content)

	if task.ReplyFn != nil {
		log.Debug("replay message to line ...")
		reply, urls := getImageUrlsFromReply(respMsg.Content)
		err = task.ReplyFn(reply, urls...)
		if err != nil {
			return errors.ErrorAt(err)
		}
	}

	return nil
}

var reImgUrl = regexp.MustCompile(`https://image.pollinations.ai/prompt/[-a-zA-Z0-9@:%_\+,.~#?&//=]+`)

func getImageUrlsFromReply(reply string) (string, []string) {
	urls := reImgUrl.FindAllString(reply, -1)
	imgUrls := make([]string, 0, len(urls))
	for i, url := range urls {
		imgUrls = append(imgUrls, url)
		reply = strings.Replace(reply, url, fmt.Sprintf("圖%d", i+1), -1)
	}

	return reply, imgUrls
}

type ClearSessionTask struct {
	SessionID string
	ReplyFn   func(reply string, imgUrls ...string) error
}

func (task *ClearSessionTask) Do(bot *Bot) error {
	log.Infof("clear session %s ...", task.SessionID)
	bot.sessMgr.GetSession(task.SessionID).Clear()

	if task.ReplyFn != nil {
		log.Debug("reply message to line ...")
		err := task.ReplyFn("已清空，小愛忘記了之前所有的對話")
		if err != nil {
			return errors.ErrorAt(err)
		}
	}
	return nil
}

type ChangeRoleTask struct {
	SessionID string
	Role      string
	ReplyFn   func(reply string, imgUrls ...string) error
}

func (task *ChangeRoleTask) Do(bot *Bot) error {
	// task.Role = must.Value(must.Value(gocc.New("s2tw")).Convert(task.Role))
	_, ok := bot.cfg.Roles[task.Role]
	var msg string
	if ok {
		log.Infof("session(%s) 變更角色為<%s>", task.SessionID, task.Role)
		bot.sessMgr.GetSession(task.SessionID).ChangeRole(task.Role)
		msg = fmt.Sprintf("小愛將扮演<%s>", task.Role)
	} else {
		keys := maps.Keys(bot.cfg.Roles)
		slices.Sort(keys)
		msg = fmt.Sprintf("角色不存在。\n您可以指定小愛扮演的角色如下:\n%s", strings.Join(keys, "\n"))
	}

	if task.ReplyFn != nil {
		log.Debug("reply message to line ...")
		err := task.ReplyFn(msg)
		if err != nil {
			return errors.ErrorAt(err)
		}
	}
	return nil
}

type ClearExpiredSessionsTask struct {
	PushMessageFn func(sessionID, msg string) error
}

func (task *ClearExpiredSessionsTask) Do(bot *Bot) error {
	ids := bot.sessMgr.ClearExpiredSessions(bot.cfg.SessionExpirePeriod)
	if len(ids) == 0 {
		return nil
	}
	log.Infof("clear expired sessions: %v", ids)

	if task.PushMessageFn == nil {
		return nil
	}

	msg := fmt.Sprintf(`超過 %v 沒有人跟小愛說話了，小愛將忘記剛剛所有的對話`, bot.cfg.SessionExpirePeriod)

	var errs error
	for _, id := range ids {
		log.Debug("push expired session notify ...")
		err := task.PushMessageFn(id, msg)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}
