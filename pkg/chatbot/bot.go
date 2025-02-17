package chatbot

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jopbrown/gobase/errors"
	"github.com/jopbrown/gobase/log"
	"github.com/jopbrown/gptbot/pkg/cfgs"
	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/sashabaranov/go-openai"
)

type Bot struct {
	cfg         *cfgs.Config
	gptClient   *openai.Client
	lineClients map[string]*linebot.Client

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
	bot.userNameCache = make(map[string]string)

	bot.lineClients = make(map[string]*linebot.Client, len(cfg.Bots))
	for path, botcfg := range cfg.Bots {
		client, err := linebot.New(botcfg.LineChannelSecret, botcfg.LineChannelToken)
		if err != nil {
			return nil, errors.ErrorAt(err)
		}
		bot.lineClients[path] = client

		botInfo, err := client.GetBotInfo().Do()
		if err != nil {
			return nil, errors.ErrorAt(err)
		}
		bot.userNameCache[path] = botInfo.DisplayName
	}

	gptCfg := openai.DefaultConfig(cfg.ChatGptAccessToken)
	gptCfg.BaseURL = cfg.ChatGptApiUrl
	bot.gptClient = openai.NewClientWithConfig(gptCfg)

	bot.sessMgr = NewSessionManager()
	bot.taskQueue = make(chan Task, bot.cfg.MaxTaskQueueCap)

	bot.handler = gin.Default()
	err = bot.registerRoute()
	if err != nil {
		return nil, errors.ErrorAt(err)
	}

	bot.stop = make(chan struct{})

	if !cfg.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	return bot, nil
}

func (bot *Bot) Serve() error {
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

	for {
		select {
		case <-ticker.C:
			ids := bot.sessMgr.ClearExpiredSessions(bot.cfg.SessionExpirePeriod)
			if len(ids) != 0 {
				log.Infof("clear expired sessions: %v", ids)
			}
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
	for key := range bot.cfg.Bots {
		bot.handler.POST(key, bot.linebotCallback)
	}
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
