package discord

import (
	"EventBot/internal/data"
	"EventBot/log"
	"fmt"
	"slices"
	"strings"

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
		d.NewOnArchiveCategorySet,
		d.NewOnChannelNameFormatSet,
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
					Flags: dg.MessageFlagsEphemeral,
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
			}
			return
		}
		err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
			Type: dg.InteractionResponseChannelMessageWithSource,
			Data: &dg.InteractionResponseData{
				Flags: dg.MessageFlagsEphemeral,
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
					Flags: dg.MessageFlagsEphemeral,
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
				Flags: dg.MessageFlagsEphemeral,
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
					Flags: dg.MessageFlagsEphemeral,
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
				Flags: dg.MessageFlagsEphemeral,
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
			{
				Type:        dg.ApplicationCommandOptionBoolean,
				Name:        "at",
				Description: "全員をAtする",
				Required:    false,
			},
			{
				Type:        dg.ApplicationCommandOptionBoolean,
				Name:        "new-channel",
				Description: "同じカテゴリーに新しいチャンネルを作成し、今のチャンネルをアーカイブに移動する（デフォルト: ON）",
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
		var title string
		at := false
		newChannel := true
		for _, o := range i.ApplicationCommandData().Options {
			switch o.Name {
			case "title":
				title = o.StringValue()
			case "at":
				at = o.BoolValue()
			case "new-channel":
				newChannel = o.BoolValue()
			}
		}
		archive := ""
		nameFormat := ""
		d.data.Get(i.GuildID, func(c *data.Content) {
			if title == "" {
				title = c.DefaultEvent
			}
			archive = c.ArchiveCategory
			nameFormat = c.ChannelNameFormat
		})
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
		if at {
			e.Description = "@everyone\n"
		}
		var m *dg.Message
		if newChannel {
			m = d.setEventInNewChannel(s, i, e, archive, strings.ReplaceAll(nameFormat, "$title", title))
			if m == nil {
				return
			}
		} else {
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
			m, err = s.InteractionResponse(i.Interaction)
			if err != nil {
				log.Error("get interaction failed", zap.Error(err))
				return
			}
		}
		d.data.Get(i.GuildID, func(c *data.Content) {
			for _, v := range c.Category {
				err := s.MessageReactionAdd(m.ChannelID, m.ID, v.Emoji)
				if err != nil {
					log.Error("Add reaction failed", zap.Error(err))
					continue
				}
			}
		})
	}
}

// setEventInNewChannel creates a copy of the current channel in the same category,
// posts the event there and moves the current channel to the archive category (if set).
// name is the name of the new channel; when empty the current channel name is reused.
// It returns the event message, or nil when the event could not be posted.
func (d *Discord) setEventInNewChannel(s *dg.Session, i *dg.InteractionCreate, e *dg.MessageEmbed, archive, name string) *dg.Message {
	// Channel operations may take longer than the 3s interaction deadline.
	err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
		Type: dg.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Error("Send message failed", zap.Error(err))
		return nil
	}
	result := func(title, desc string) {
		_, err := s.InteractionResponseEdit(i.Interaction, &dg.WebhookEdit{
			Embeds: &[]*dg.MessageEmbed{
				{
					Title:       title,
					Description: desc,
					Color:       0xFEDFE1,
				},
			},
		})
		if err != nil {
			log.Error("Edit interaction response failed", zap.Error(err))
		}
	}
	cur, err := s.Channel(i.ChannelID)
	if err != nil {
		log.Error("Get channel failed", zap.Error(err))
		result("失敗しました", "チャンネル情報を取得できませんでした")
		return nil
	}
	if cur.Type != dg.ChannelTypeGuildText {
		result("失敗しました", "このチャンネルでは新しいチャンネルを作成できません。`new-channel: False` を指定してください")
		return nil
	}
	if name == "" {
		name = cur.Name
	} else if r := []rune(name); len(r) > 100 {
		// Discord limits channel names to 100 characters.
		name = string(r[:100])
	}
	created, err := s.GuildChannelCreateComplex(i.GuildID, dg.GuildChannelCreateData{
		Name:                 name,
		Type:                 cur.Type,
		Topic:                cur.Topic,
		RateLimitPerUser:     cur.RateLimitPerUser,
		Position:             cur.Position,
		PermissionOverwrites: cur.PermissionOverwrites,
		ParentID:             cur.ParentID,
		NSFW:                 cur.NSFW,
	})
	if err != nil {
		log.Error("Create channel failed", zap.Error(err))
		result("失敗しました", "新しいチャンネルを作成できませんでした。Botの権限（チャンネルの管理）を確認してください")
		return nil
	}
	m, err := s.ChannelMessageSendEmbed(created.ID, e)
	if err != nil {
		log.Error("Send message failed", zap.Error(err))
		result("失敗しました", fmt.Sprintf("<#%s> を作成しましたが、イベントを送信できませんでした", created.ID))
		return nil
	}
	desc := fmt.Sprintf("新しいチャンネル <#%s> にイベントを設定しました。", created.ID)
	if archive != "" && archive != cur.ParentID {
		_, err = s.ChannelEditComplex(cur.ID, &dg.ChannelEdit{ParentID: archive})
		if err != nil {
			log.Error("Move channel failed", zap.Error(err))
			desc += "\nこのチャンネルをアーカイブカテゴリーに移動できませんでした。"
		}
	}
	result("設定しました", desc)
	return m
}

func (d *Discord) NewOnArchiveCategorySet() (*dg.ApplicationCommand, func(s *dg.Session, i *dg.InteractionCreate)) {
	cmd := &dg.ApplicationCommand{
		Name:        "set-archive-category",
		Description: "set-event時に古いチャンネルを移動するカテゴリーを設定します（未指定で解除）",
		Options: []*dg.ApplicationCommandOption{
			{
				Type:         dg.ApplicationCommandOptionChannel,
				Name:         "category",
				Description:  "移動先のカテゴリー",
				ChannelTypes: []dg.ChannelType{dg.ChannelTypeGuildCategory},
				Required:     false,
			},
		},
	}
	return cmd, func(s *dg.Session, i *dg.InteractionCreate) {
		if len(i.ChannelID) == 0 {
			return
		}
		id := ""
		for _, o := range i.ApplicationCommandData().Options {
			if o.Name == "category" {
				id = o.ChannelValue(nil).ID
			}
		}
		d.data.Set(i.GuildID, func(c *data.Content) {
			c.ArchiveCategory = id
		})
		e := &dg.MessageEmbed{
			Title:       "設定しました",
			Description: fmt.Sprintf("アーカイブカテゴリーを <#%s> に設定しました", id),
			Color:       0xFEDFE1,
		}
		if id == "" {
			e.Description = "アーカイブカテゴリーを解除しました"
		}
		err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
			Type: dg.InteractionResponseChannelMessageWithSource,
			Data: &dg.InteractionResponseData{
				Flags: dg.MessageFlagsEphemeral,
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
				Flags: dg.MessageFlagsEphemeral,
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

func (d *Discord) NewOnChannelNameFormatSet() (*dg.ApplicationCommand, func(s *dg.Session, i *dg.InteractionCreate)) {
	cmd := &dg.ApplicationCommand{
		Name:        "set-channel-name-format",
		Description: "set-eventで作成するチャンネル名のフォーマットを設定します（未指定で解除）",
		Options: []*dg.ApplicationCommandOption{
			{
				Type:        dg.ApplicationCommandOptionString,
				Name:        "format",
				Description: "チャンネル名のフォーマット。$title はイベントのタイトルに置き換えられます",
				MaxLength:   100,
				Required:    false,
			},
		},
	}
	return cmd, func(s *dg.Session, i *dg.InteractionCreate) {
		if len(i.ChannelID) == 0 {
			return
		}
		format := ""
		for _, o := range i.ApplicationCommandData().Options {
			if o.Name == "format" {
				format = strings.TrimSpace(o.StringValue())
			}
		}
		d.data.Set(i.GuildID, func(c *data.Content) {
			c.ChannelNameFormat = format
		})
		e := &dg.MessageEmbed{
			Title:       "設定しました",
			Description: fmt.Sprintf("チャンネル名のフォーマットを `%s` に設定しました", format),
			Color:       0xFEDFE1,
		}
		if format == "" {
			e.Description = "チャンネル名のフォーマットを解除しました。今のチャンネルと同じ名前で作成します"
		}
		err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
			Type: dg.InteractionResponseChannelMessageWithSource,
			Data: &dg.InteractionResponseData{
				Flags: dg.MessageFlagsEphemeral,
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
