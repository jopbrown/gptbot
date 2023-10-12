package main

import (
	"path/filepath"
	"time"

	"github.com/jopbrown/gobase/errors"
	"github.com/jopbrown/gobase/fsutil"
	"github.com/jopbrown/gobase/log"
	"github.com/jopbrown/gobase/log/rotate"
	"github.com/jopbrown/gptbot/pkg/cfgs"
	"github.com/jopbrown/gptbot/pkg/chatbot"
)

var (
	BuildName    = "myapp"
	BuildVersion = "v0.0.0"
	BuildHash    = "unknown"
	BuildTime    = "20060102150405"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(errors.GetErrorDetails(err))
	}
}

func run() error {
	cfg, err := cfgs.LoadConfig(filepath.Join(fsutil.AppDir(), "gptbot.yaml"))
	if err != nil {
		return errors.ErrorAt(err)
	}
	cfg.MergeDefault()

	err = applyLog(cfg)
	if err != nil {
		return errors.ErrorAt(err)
	}
	cfg.WriteConfig(log.GetWriter(log.LevelDebug))

	log.Infof("%s %v-%v-%v", BuildName, BuildVersion, BuildHash, BuildTime)

	bot, err := chatbot.NewBot(cfg)
	if err != nil {
		return errors.ErrorAt(err)
	}

	err = bot.Serve()
	if err != nil {
		return errors.ErrorAt(err)
	}

	return nil
}

func applyLog(cfg *cfgs.Config) error {
	f, err := rotate.OpenFile(filepath.Join(cfg.LogPath, "gptbot.log"), 24*time.Hour, 0)
	if err != nil {
		return errors.ErrorAt(err)
	}

	tee := log.NewTeeLogger(
		log.ConsoleLogger(cfg.DebugMode),
		log.FileLogger(f, log.FileLoggerFormat(), cfg.DebugMode),
	)

	log.SetGlobalLogger(tee)

	return nil
}
