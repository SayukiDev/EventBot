package discord

import (
	"EventBot/internal/data"
	"EventBot/log"
	"fmt"
	"strings"

	dg "github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func (d *Discord) OnReactionUpdate(s *dg.Session, gID, cId, mId string) {
	var fields []*dg.MessageEmbedField
	var title string
	m, err := s.ChannelMessage(cId, mId)
	if err != nil {
		log.Error("Get message failed", zap.Error(err))
		return
	}
	if m.Author.ID != s.State.User.ID {
		return
	}
	if len(m.Embeds) == 0 {
		return
	}
	if strings.HasPrefix(m.Embeds[0].Title, ":calendar_spiral:") {
		title = m.Embeds[0].Title
	} else {
		return
	}
	d.data.Get(gID, func(c *data.Content) {
		for _, v := range c.Category {
			us, err := s.MessageReactions(cId, mId, v.Emoji, 100, "", "")
			if err != nil {
				log.Error("Get reactions failed", zap.Error(err))
				return
			}
			list := fmt.Sprintf("計 %d 人\n", len(us)-1)
			for _, u := range us {
				if u.Bot {
					continue
				}
				list += "\n" + u.DisplayName()
			}
			list += "\n\n-------------\n\n"
			fields = append(fields, &dg.MessageEmbedField{
				Name:  fmt.Sprintf("%s **%s**", v.Emoji, v.Name),
				Value: list,
			})
		}
	})
	e := &dg.MessageEmbed{
		Title:  title,
		Fields: fields,
		Color:  0xFEDFE1,
	}
	_, err = s.ChannelMessageEditEmbed(cId, mId, e)
	if err != nil {
		log.Error("Edit message failed", zap.Error(err))
		return
	}
}

func (d *Discord) OnReactionAdd(s *dg.Session, r *dg.MessageReactionAdd) {
	if d.s.State.User.ID == r.UserID {
		return
	}
	log.Info("ReactionAdd",
		zap.String("emoji", r.Emoji.Name),
		zap.String("channel", r.ChannelID),
		zap.String("message", r.MessageID))
	d.OnReactionUpdate(s, r.GuildID, r.ChannelID, r.MessageID)
}

func (d *Discord) OnReactionRemove(s *dg.Session, r *dg.MessageReactionRemove) {
	if d.s.State.User.ID == r.UserID {
		return
	}
	log.Info("ReactionRemove",
		zap.String("emoji", r.Emoji.Name),
		zap.String("channel", r.ChannelID),
		zap.String("message", r.MessageID))
	d.OnReactionUpdate(s, r.GuildID, r.ChannelID, r.MessageID)
}

func (d *Discord) AddHandle() error {
	d.s.AddHandler(d.OnReactionAdd)
	d.s.AddHandler(d.OnReactionRemove)
	return nil
}
