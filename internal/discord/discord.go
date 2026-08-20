package discord

import (
	"EventBot/internal/data"
	"fmt"
	"time"

	dg "github.com/bwmarrin/discordgo"
)

type Discord struct {
	s          *dg.Session
	data       *data.Data
	cmdHandles map[string]func(s *dg.Session, i *dg.InteractionCreate)
}

func NewDiscord(data *data.Data, token string) (*Discord, error) {
	s, err := dg.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	return &Discord{
		s:          s,
		data:       data,
		cmdHandles: make(map[string]func(s *dg.Session, i *dg.InteractionCreate)),
	}, nil
}

func (d *Discord) Start() error {
	d.s.Identify.Intents =
		dg.IntentsGuildMessages |
			dg.IntentMessageContent |
			dg.IntentGuildMessageReactions |
			dg.IntentsGuildMessageReactions
	err := d.s.Open()
	if err != nil {
		return err
	}
	err = d.AddHandle()
	if err != nil {
		return fmt.Errorf("add handle error: %w", err)
	}
	err = d.AddCommand()
	if err != nil {
		return fmt.Errorf("add command error: %w", err)
	}
	go func() {
		for ; ; time.Sleep(time.Second * 360) {
			d.s.UpdateCustomStatus("Running...")
		}
	}()
	return nil
}
