package discord

import (
	"EventBot/internal/data"
	"EventBot/log"
	"fmt"
	"slices"

	dg "github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func (d *Discord) AddCommand() error {
	newer := []cmdNewer{
		d.NewOnCategoryList,
		d.NewOnCategoryAdd,
		d.NewOnCategoryDel,
		d.NewOnEventSet,
		d.NewOnDefaultEventSet,
	}
	for _, v := range newer {
		cmd, h := v()
		_, err := d.s.ApplicationCommandCreate(d.s.State.User.ID, "", cmd)
		if err != nil {
			return fmt.Errorf("create OnCategoryAdd error: %w", err)
		}
		d.cmdHandles[cmd.Name] = h
	}
	d.s.AddHandler(func(s *dg.Session, i *dg.InteractionCreate) {
		log.Info("InteractionCreate", zap.String("type", i.Type.String()), zap.String("name", i.ApplicationCommandData().Name))
		if h, ok := d.cmdHandles[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
	return nil
}

type cmdNewer func() (*dg.ApplicationCommand, func(s *dg.Session, i *dg.InteractionCreate))

func (d *Discord) NewOnCategoryAdd() (*dg.ApplicationCommand, func(s *dg.Session, i *dg.InteractionCreate)) {
	cmd := &dg.ApplicationCommand{
		Name:        "add-category",
		Description: "カテゴリーを追加",
		Options: []*dg.ApplicationCommandOption{
			{
				Type:        dg.ApplicationCommandOptionString,
				Name:        "emoji",
				Description: "絵文字",
				Required:    true,
			},
			{
				Type:        dg.ApplicationCommandOptionString,
				Name:        "name",
				Description: "絵文字に対応する名前",
				Required:    true,
			},
		},
	}
	return cmd, func(s *dg.Session, i *dg.InteractionCreate) {
		if len(i.ChannelID) == 0 {
			return
		}
		ad := i.ApplicationCommandData()
		if len(ad.Options) < 2 {
			log.Error("No option")
			return
		}
		emoji := ad.Options[0].StringValue()
		name := ad.Options[1].StringValue()
		failed := false
		d.data.Set(i.GuildID, func(c *data.Content) {
			for _, v := range c.Category {
				if v.Emoji == emoji {
					log.Error("Category already exists", zap.String("emoji", emoji))
					failed = true
					return
				}
			}
			c.Category = append(c.Category, data.Category{Emoji: emoji, Name: name})
		})
		if failed {
			err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
				Type: dg.InteractionResponseChannelMessageWithSource,
				Data: &dg.InteractionResponseData{
					Embeds: []*dg.MessageEmbed{
						{
							Title:       "失敗しました",
							Description: "カテゴリー既に存在しています",
							Color:       0xFEDFE1,
						},
					},
				},
			})
			if err != nil {
				log.Error("Send message failed", zap.Error(err))
				return
			}
		}
		err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
			Type: dg.InteractionResponseChannelMessageWithSource,
			Data: &dg.InteractionResponseData{
				Embeds: []*dg.MessageEmbed{
					{
						Title:       "追加しました。",
						Description: "カテゴリーを追加しました。",
						Color:       0xFEDFE1,
					},
				},
			},
		})
		if err != nil {
			log.Error("Send message failed", zap.Error(err))
			return
		}
	}
}

func (d *Discord) NewOnCategoryList() (*dg.ApplicationCommand, func(s *dg.Session, i *dg.InteractionCreate)) {
	cmd := dg.ApplicationCommand{
		Name:        "list-category",
		Description: "カテゴリーのリストを取得する",
	}
	return &cmd, func(s *dg.Session, i *dg.InteractionCreate) {
		if len(i.ChannelID) == 0 {
			return
		}
		var fields []*dg.MessageEmbedField
		ok := d.data.Get(i.GuildID, func(c *data.Content) {
			for _, v := range c.Category {
				fields = append(fields, &dg.MessageEmbedField{
					Name:  v.Emoji,
					Value: v.Name + "\n",
				})
			}
		})
		if !ok || len(fields) == 0 {
			err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
				Type: dg.InteractionResponseChannelMessageWithSource,
				Data: &dg.InteractionResponseData{
					Embeds: []*dg.MessageEmbed{
						{
							Title:       "取得に失敗しました",
							Description: "カテゴリー存在していません",
							Color:       0xFEDFE1,
						},
					},
				},
			})
			if err != nil {
				log.Error("Send message failed", zap.Error(err))
			}
			return
		}
		err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
			Type: dg.InteractionResponseChannelMessageWithSource,
			Data: &dg.InteractionResponseData{
				Embeds: []*dg.MessageEmbed{
					{
						Title:  "取得しました",
						Fields: fields,
						Color:  0xFEDFE1,
					},
				},
			},
		})
		if err != nil {
			log.Error("Send message failed", zap.Error(err))
			return
		}
	}
}

func (d *Discord) NewOnCategoryDel() (*dg.ApplicationCommand, func(s *dg.Session, i *dg.InteractionCreate)) {
	cmd := dg.ApplicationCommand{
		Name:        "del-category",
		Description: "かりごりーを削除",
		Options: []*dg.ApplicationCommandOption{
			{
				Type:        dg.ApplicationCommandOptionString,
				Name:        "emoji",
				Description: "絵文字",
				Required:    true,
			},
		},
	}
	return &cmd, func(s *dg.Session, i *dg.InteractionCreate) {
		if len(i.ChannelID) == 0 {
			return
		}
		ad := i.ApplicationCommandData()
		if len(ad.Options) == 0 {
			log.Error("No option")
			return
		}
		emoji := ad.Options[0].StringValue()
		ok := false
		d.data.Set(i.GuildID, func(c *data.Content) {
			for in, v := range c.Category {
				if v.Emoji != emoji {
					continue
				}
				ok = true
				c.Category = slices.Delete(c.Category, in, in+1)
				return
			}
		})
		if ok {
			err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
				Type: dg.InteractionResponseChannelMessageWithSource,
				Data: &dg.InteractionResponseData{
					Embeds: []*dg.MessageEmbed{
						{
							Title:       "削除しました。",
							Description: "カテゴリーを削除しました。",
							Color:       0xFEDFE1,
						},
					},
				},
			})
			if err != nil {
				log.Error("Send message failed", zap.Error(err))
				return
			}
			return
		}
		err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
			Type: dg.InteractionResponseChannelMessageWithSource,
			Data: &dg.InteractionResponseData{
				Embeds: []*dg.MessageEmbed{
					{
						Title:       "失敗しました",
						Description: "カテゴリー存在していません",
						Color:       0xFEDFE1,
					},
				},
			},
		})
		if err != nil {
			log.Error("Send message failed", zap.Error(err))
			return
		}
	}
}

func (d *Discord) NewOnEventSet() (*dg.ApplicationCommand, func(s *dg.Session, i *dg.InteractionCreate)) {
	cmd := &dg.ApplicationCommand{
		Name:        "set-event",
		Description: "イベントを設定します",
		Options: []*dg.ApplicationCommandOption{
			{
				Type:        dg.ApplicationCommandOptionString,
				Name:        "title",
				Description: "イベントのタイトル",
				Required:    false,
			},
		},
	}
	return cmd, func(s *dg.Session, i *dg.InteractionCreate) {
		if len(i.ChannelID) == 0 {
			return
		}
		var fields []*dg.MessageEmbedField
		ok := d.data.Get(i.GuildID, func(c *data.Content) {
			for _, v := range c.Category {
				fields = append(fields, &dg.MessageEmbedField{
					Name:  v.Emoji + v.Name,
					Value: "",
				})
			}
		})
		if !ok {
			err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
				Type: dg.InteractionResponseChannelMessageWithSource,
				Data: &dg.InteractionResponseData{
					Embeds: []*dg.MessageEmbed{
						{
							Title:       "カテゴリーが設定されていません",
							Description: "まずカテゴリーを設定してください",
							Color:       0xFEDFE1,
						},
					},
				},
			})
			if err != nil {
				log.Error("Send message failed", zap.Error(err))
				return
			}
			return
		}
		ad := i.ApplicationCommandData()
		var title string
		if len(ad.Options) == 0 {
			d.data.Get(i.GuildID, func(c *data.Content) {
				title = c.DefaultEvent
			})
		} else {
			title = ad.Options[0].StringValue()
		}
		if title == "" {
			err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
				Type: dg.InteractionResponseChannelMessageWithSource,
				Data: &dg.InteractionResponseData{
					Embeds: []*dg.MessageEmbed{
						{
							Title:       "デフォルトイベント設定されてません",
							Description: "デフォルトイベントを設定するかタイトルを指定してください",
							Color:       0xFEDFE1,
						},
					},
				},
			})
			if err != nil {
				log.Error("Send message failed", zap.Error(err))
				return
			}
			return
		}
		e := &dg.MessageEmbed{
			Title:  fmt.Sprintf(":calendar_spiral: %s\n------------", title),
			Fields: fields,
			Color:  0xFEDFE1,
		}
		err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
			Type: dg.InteractionResponseChannelMessageWithSource,
			Data: &dg.InteractionResponseData{
				Embeds: []*dg.MessageEmbed{
					e,
				},
			},
		})
		if err != nil {
			log.Error("Send message failed", zap.Error(err))
			return
		}
		m, err := s.InteractionResponse(i.Interaction)
		if err != nil {
			log.Error("get interaction failed", zap.Error(err))
			return
		}
		d.data.Set(i.GuildID, func(c *data.Content) {
			for _, v := range c.Category {
				err := s.MessageReactionAdd(i.ChannelID, m.ID, v.Emoji)
				if err != nil {
					log.Error("Add reaction failed", zap.Error(err))
					continue
				}
			}
		})
	}
}

func (d *Discord) NewOnDefaultEventSet() (*dg.ApplicationCommand, func(s *dg.Session, i *dg.InteractionCreate)) {
	cmd := &dg.ApplicationCommand{
		Name:        "set-default-event",
		Description: "イベントを設定します",
		Options: []*dg.ApplicationCommandOption{
			{
				Type:        dg.ApplicationCommandOptionString,
				Name:        "title",
				Description: "イベントのタイトル",
				Required:    true,
			},
		},
	}
	return cmd, func(s *dg.Session, i *dg.InteractionCreate) {
		if len(i.ChannelID) == 0 {
			return
		}
		ad := i.ApplicationCommandData()
		if len(ad.Options) == 0 {
			log.Error("No option")
			return
		}
		title := ad.Options[0].StringValue()
		e := &dg.MessageEmbed{
			Title:       "設定しました",
			Description: "デフォルトイベントを設定しました",
			Color:       0xFEDFE1,
		}
		d.data.Set(i.GuildID, func(c *data.Content) {
			c.DefaultEvent = title
		})
		err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
			Type: dg.InteractionResponseChannelMessageWithSource,
			Data: &dg.InteractionResponseData{
				Embeds: []*dg.MessageEmbed{
					e,
				},
			},
		})
		if err != nil {
			log.Error("Send message failed", zap.Error(err))
			return
		}
	}
}
