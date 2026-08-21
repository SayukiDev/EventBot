package main

import (
	"EventBot/config"
	"EventBot/internal/data"
	"EventBot/internal/discord"
	"EventBot/log"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"go.uber.org/zap"
)

const version = "0.0.1"

var c = flag.String("config", "./config.json", "config file path")

func main() {
	flag.Parse()
	c := config.NewConfig()
	err := c.Load("./config.json")
	if err != nil {
		log.ErrorE("Failed to load config", zap.Error(err))
	}
	log.SetLogLevel(c.LogLevel)
	log.Info(
		"Starting EventBot",
		zap.String("version", "0.0.1"),
		zap.String("go_version", runtime.Version()),
		zap.String("os", runtime.GOOS),
		zap.String("arch", runtime.GOARCH),
		zap.String("log_level", c.LogLevel),
	)
	d := data.New(c.DataPath)
	err = d.Load()
	if err != nil {
		log.ErrorE("Failed to load data", zap.Error(err))
	}
	err = d.Start()
	if err != nil {
		log.ErrorE("Failed to start data", zap.Error(err))
	}
	log.AddCloseFunc(d.Close)
	dc, err := discord.NewDiscord(d, c.Token)
	if err != nil {
		log.ErrorE("Failed to start discord", zap.Error(err))
	}
	err = dc.Start()
	if err != nil {
		log.ErrorE("Failed to start discord", zap.Error(err))
	}
	runtime.GC()
	log.Info("EventBot started.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Info("Shutting down...")
	log.Close()
}
