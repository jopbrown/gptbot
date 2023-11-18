package chatbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jopbrown/gobase/errors"
	"github.com/jopbrown/gobase/log"
	"github.com/jopbrown/gptbot/pkg/cfgs"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/sashabaranov/go-openai"
)

type Bot struct {
	cfg        *cfgs.Config
	lineClient *linebot.Client
	gptClient  *openai.Client

	sessMgr       *SessionManager
	taskQueue     chan Task
	handler       *gin.Engine
	stop          chan struct{}
	userNameCache map[string]string
}

func NewBot(cfg *cfgs.Config) (*Bot, error) {
	var err error
	bot := &Bot{}
	bot.cfg = cfg

	bot.lineClient, err = linebot.New(cfg.LineChannelSecret, cfg.LineChannelToken)
	if err != nil {
		return nil, errors.ErrorAt(err)
	}

	gptCfg := openai.DefaultConfig(cfg.ChatGptAccessToken)
	gptCfg.BaseURL = cfg.ChatGptApiUrl
	bot.gptClient = openai.NewClientWithConfig(gptCfg)

	bot.sessMgr = NewSessionManager(cfg.DefaultRole)
	bot.taskQueue = make(chan Task, bot.cfg.MaxTaskQueueCap)

	bot.handler = gin.Default()
	err = bot.registerRoute()
	if err != nil {
		return nil, errors.ErrorAt(err)
	}

	bot.stop = make(chan struct{})

	bot.userNameCache = make(map[string]string)

	if !cfg.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	return bot, nil
}

func (bot *Bot) UpdateApiServerAccessToken() error {
	type secret struct {
		Token string `json:"token"`
		PUID  string `json:"puid"`
	}

	url, err := url.ParseRequestURI(bot.cfg.ChatGptApiUrl)
	if err != nil {
		return errors.ErrorAt(err)
	}

	reqUrl := fmt.Sprintf("%s://%s/admin/tokens", url.Scheme, url.Host)
	tokens := map[string]secret{"": {Token: bot.cfg.ChatGptAccessToken}}

	log.Debugf("send update token request to `%s`", reqUrl)

	payload, err := json.Marshal(tokens)
	if err != nil {
		return errors.ErrorAt(err)
	}

	req, err := http.NewRequest("PATCH", reqUrl, bytes.NewBuffer(payload))
	if err != nil {
		return errors.ErrorAt(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "TotallySecurePassword")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if err != nil {
			return errors.ErrorAt(err)
		}

	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("update token request got non-200 status: %d", resp.StatusCode)
	}

	return nil
}

func (bot *Bot) Serve() error {

	// Verify if the service originates from the ChatGPT web application.
	if strings.HasPrefix(bot.cfg.ChatGptAccessToken, "eyJhbGciOiJSUzI1NiI") {
		err := bot.UpdateApiServerAccessToken()
		if err != nil {
			return errors.ErrorAt(err)
		}
	}

	addr := fmt.Sprintf(":%d", bot.cfg.ServePort)
	server := &http.Server{Addr: addr, Handler: bot.handler}

	go func() {
		log.Infof("gptbot serve on %s", addr)
		err := server.ListenAndServe()
		if err != nil {
			log.Errorf("fail to ListenAndServe: %v", err)
		}
		bot.Stop()
	}()

	go bot.DoTasks()
	go bot.ClearExpiredSessionsPeriodically()

	<-bot.stop
	log.Info("gptbot stop serve")

	return nil
}

func (bot *Bot) DoTasks() {
	for {
		select {
		case task := <-bot.taskQueue:
			err := task.Do(bot)
			if err != nil {
				log.ErrorAt(err)
			}
		case <-bot.stop:
			return
		}
	}
}

func (bot *Bot) ClearExpiredSessionsPeriodically() {
	ticker := time.NewTicker(bot.cfg.SessionClearInterval)
	defer ticker.Stop()

	task := &ClearExpiredSessionsTask{}
	if !bot.cfg.NotPushExpireMessage {
		task.PushMessageFn = func(sessionID, msg string) error {
			_, err := bot.lineClient.PushMessage(sessionID, linebot.NewTextMessage(msg)).Do()
			if err != nil {
				return errors.ErrorAt(err)
			}
			return nil
		}
	}

	for {
		select {
		case <-ticker.C:
			bot.taskQueue <- task
		case <-bot.stop:
			return
		}
	}
}

func (bot *Bot) Stop() {
	close(bot.stop)
}

func (bot *Bot) registerRoute() error {
	bot.handler.GET("/ping", bot.pingHandler)
	bot.handler.GET("/stop", bot.stopHandler)
	bot.handler.POST("/linebot", bot.linebotCallback)
	return nil
}

func (bot *Bot) pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func (bot *Bot) stopHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "stop service after 5s",
	})
	log.Info("stop service after 5s")
	time.Sleep(5 * time.Second)
	bot.Stop()
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
