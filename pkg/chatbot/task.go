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
	UserName string
	Session  *Session
	Message  string
	IsGroup  bool
	ReplyFn  func(reply string, imgUrls ...string) error
}

func (task *ChatTask) Do(bot *Bot) error {
	log.Debugf("do chat task...\n %+v", task)

	recorder, err := rotate.OpenFile(filepath.Join(bot.cfg.LogPath, "chats", fmt.Sprintf("%s.txt", task.Session.ID)), 24*time.Hour, 0)
	if err != nil {
		return errors.ErrorAt(err)
	}
	defer recorder.Close()

	role := bot.cfg.Roles[task.Session.Role]
	if role.MaxConversationCount > 0 && len(task.Session.Messages) >= role.MaxConversationCount*2+1 {
		task.Session.Clear()
	}

	cmds := bot.cfg.CmdsTalkToAI
	if len(role.CmdsTalkToAI) > 0 {
		cmds = role.CmdsTalkToAI
	}
	msg := task.Message
	msg, isTalkToAI := messageMatchCmd(msg, cmds)
	if task.IsGroup && !isTalkToAI && !role.NotNeedSlashCmd {
		log.Debug("skip talk to AI")
		return nil
	}

	if len(task.Session.Messages) == 0 && len(role.Prompt) != 0 {
		log.Debug("append system message ...")
		task.Session.AddMessage(&openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: role.Prompt,
		})
		fmt.Fprintln(recorder, "#####################")
		fmt.Fprintln(recorder, role.Prompt)
	}

	toMsg := fmt.Sprintf("%s: %s", task.UserName, msg)
	if role.PrefixUserName {
		msg = toMsg
	}
	log.Info(msg)
	fmt.Fprintln(recorder, toMsg)

	chatMsg := &openai.ChatCompletionMessage{}
	chatMsg.Content = msg
	chatMsg.Role = openai.ChatMessageRoleUser
	task.Session.AddMessage(chatMsg)

	log.Debug("send message to chatgpt ...")
	resp, err := bot.gptClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    bot.cfg.ChatGptModel,
			Messages: task.Session.Messages,
		},
	)

	if err != nil {
		if task.ReplyFn != nil {
			var err1 error
			switch GetOpenAIErrCode(err) {
			case 401:
				err1 = task.ReplyFn(fmt.Sprintf("AI的token過期了，請聯繫管理員更新:\n%s", errors.GetErrorDetails(err)))
			case 500:
				err1 = task.ReplyFn(fmt.Sprintf("Server掛掉了，請聯繫管理員:\n%s", errors.GetErrorDetails(err)))
			default:
				err1 = task.ReplyFn(fmt.Sprintf("小愛壞掉了，可以嘗試輸入清空指令修復:\n%s", errors.GetErrorDetails(err)))
			}

			if err1 != nil {
				return errors.ErrorAt(errors.Join(err, err1))
			}
		}
		return errors.ErrorAt(err)
	}

	respMsg := &openai.ChatCompletionMessage{}
	respMsg.Role = openai.ChatMessageRoleAssistant
	respMsg.Content = resp.Choices[0].Message.Content

	task.Session.AddMessage(respMsg)
	log.Info("AI:", respMsg.Content)
	fmt.Fprintln(recorder, "AI:", respMsg.Content)

	if task.ReplyFn != nil {
		log.Debug("replay message to line ...")
		reply, urls := getImageUrlsFromReply(respMsg.Content)
		reply = strings.TrimSpace(reply)
		err = task.ReplyFn(reply, urls...)
		if err != nil {
			return errors.ErrorAt(err)
		}
	}

	return nil
}

var reImgUrl = regexp.MustCompile(`https://image.pollinations.ai/prompt/[-a-zA-Z0-9@:%_\+,.~#?&//=\s]+`)

func getImageUrlsFromReply(reply string) (string, []string) {
	urls := reImgUrl.FindAllString(reply, -1)
	imgUrls := make([]string, 0, len(urls))
	for i, url := range urls {
		tidyUrl := strings.TrimSpace(url)
		tidyUrl = strings.ReplaceAll(tidyUrl, " ", "-")
		imgUrls = append(imgUrls, tidyUrl)
		reply = strings.Replace(reply, url, fmt.Sprintf("圖%d", i+1), -1)
	}

	return reply, imgUrls
}

type ClearSessionTask struct {
	Session *Session
	ReplyFn func(reply string, imgUrls ...string) error
}

func (task *ClearSessionTask) Do(bot *Bot) error {
	log.Infof("clear session %s ...", task.Session.ID)
	task.Session.Clear()

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
	Session *Session
	Role    string
	ReplyFn func(reply string, imgUrls ...string) error
}

func (task *ChangeRoleTask) Do(bot *Bot) error {
	_, ok := bot.cfg.Roles[task.Role]
	var msg string
	if ok {
		log.Infof("session(%s) 變更角色為<%s>", task.Session.ID, task.Role)
		task.Session.ChangeRole(task.Role)
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
