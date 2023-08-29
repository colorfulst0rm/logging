package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

type Channel struct {
	Id        string `json:"id" db:"id"`
	Snowflake string `json:"snowflake" db:"snowflake"`
	Guild     string `json:"guild" db:"guild"`
	Name      string `json:"name" db:"name"`
}

type Guild struct {
	Id        string `json:"id" db:"id"`
	Snowflake string `json:"snowflake" db:"snowflake"`
	Name      string `json:"name" db:"name"`
}

type Message struct {
	Id          string    `json:"id" db:"id"`
	Snowflake   string    `json:"snowflake" db:"snowflake"`
	Channel     string    `json:"channel" db:"channel"`
	Guild       string    `json:"guild" db:"guild"`
	Edited      time.Time `json:"edited" db:"edited"`
	Content     string    `json:"content" db:"content"`
	Reference   string    `json:"reference" db:"reference"`
	Edits       []string  `json:"edits" db:"edits"`
	Embeds      []string  `json:"embeds" db:"embeds"`
	Author      string    `json:"author" db:"author"`
	Attachments []string  `json:"attachments" db:"attachments"`
}

type Attachment struct {
	Id       string `json:"id" db:"id"`
	Url      string `json:"url" db:"url"`
	ProxyUrl string `json:"proxy_url" db:"proxy_url"`
	Filename string `json:"filename" db:"filename"`
}

type User struct {
	Id          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	DisplayName string `json:"display_name" db:"display_name"`
	Guild       string `json:"guild" db:"guild"`
	Snowflake   string `json:"snowflake" db:"snowflake"`
}

type Embed struct {
	Id            string   `json:"id" db:"id"`
	Title         string   `json:"title" db:"title"`
	Type          string   `json:"type" db:"type"`
	Url           string   `json:"url" db:"url"`
	Timestamp     string   `json:"timestamp" db:"timestamp"`
	Color         string   `json:"color" db:"color"`
	FooterText    string   `json:"footer_text" db:"footer_text"`
	FooterIconURL string   `json:"footer_icon_url" db:"footer_icon_url"`
	Video         string   `json:"video" db:"video"`
	Provider      string   `json:"provider" db:"provider"`
	Fields        []string `json:"fields" db:"fields"`
	Image         string   `json:"image" db:"image"`
	Thumbnail     string   `json:"thumbnail" db:"thumbnail"`
}

type EmbedField struct {
	Id     string `json:"id" db:"id"`
	Name   string `json:"name" db:"name"`
	Value  string `json:"value" db:"value"`
	Inline bool   `json:"inline" db:"inline"`
}

func create_channel(app *pocketbase.PocketBase, s *discordgo.Session, channelid string) (channel Channel) {
	if app == nil || s == nil {
		return
	}
	channel = Channel{}
	err := app.Dao().DB().
		Select("snowflake", "guild", "name", "id").
		From("discord_channels").
		AndWhere(dbx.In("snowflake", channelid)).
		One(&channel)

	dchannel, err2 := s.Channel(channelid)

	if err2 != nil {
		return
	}

	if err != nil {
		guild := create_guild(app, s, dchannel.GuildID)
		// create channel

		collection, err3 := app.Dao().FindCollectionByNameOrId("discord_channels")

		if err3 != nil {
			return
		}

		record := models.NewRecord(collection)
		record.Set("snowflake", dchannel.ID)
		record.Set("guild", guild.Id)
		record.Set("name", dchannel.Name)

		if err := app.Dao().SaveRecord(record); err != nil {
			return
		}

		// _, err3 := app.Dao().DB().
		// 	Insert("discord_channels", dbx.Params{
		// 		"snowflake": dchannel.ID,
		// 		"guild":     guild.Id,
		// 		"name":      dchannel.Name,
		// 	}).
		// 	Execute()
		if err3 != nil {
			return
		}
		err := app.Dao().DB().
			Select("snowflake", "guild", "name", "id").
			From("discord_channels").
			AndWhere(dbx.In("snowflake", channelid)).
			One(&channel)
		if err != nil {
			return
		}
	}
	return channel
}

func create_guild(app *pocketbase.PocketBase, s *discordgo.Session, guildid string) (guild Guild) {
	guild = Guild{}
	err := app.Dao().DB().
		Select("snowflake", "name", "id").
		From("discord_guilds").
		AndWhere(dbx.In("snowflake", guildid)).
		One(&guild)

	if err == nil {
		return guild
	}

	dguild, err2 := s.Guild(guildid)

	if err2 != nil {
		return
	}

	collection, err := app.Dao().FindCollectionByNameOrId("discord_guilds")

	if err == nil {
		// create channel

		record := models.NewRecord(collection)
		record.Set("snowflake", dguild.ID)
		record.Set("name", dguild.Name)

		if err := app.Dao().SaveRecord(record); err != nil {
			return
		}

		// _, err3 := app.Dao().DB().
		// 	Insert("discord_guilds", dbx.Params{
		// 		"snowflake": dguild.ID,
		// 		"name":      dguild.Name,
		// 	}).
		// 	Execute()
		// if err3 != nil {
		// 	return
		// }
		err := app.Dao().DB().
			Select("snowflake", "name").
			From("discord_guilds").
			AndWhere(dbx.In("snowflake", guildid)).
			One(&guild)
		if err != nil {
			return
		}
	}
	return guild
}

func get_message(app *pocketbase.PocketBase, s *discordgo.Session, m string) (message Message, err error) {
	message = Message{}
	err = app.Dao().DB().
		Select("snowflake", "channel", "guild", "edited", "content", "reference", "edits", "embeds", "author", "id").
		From("discord_messages").
		AndWhere(dbx.In("snowflake", m)).
		One(&message)

	return message, err

}

func get_user(app *pocketbase.PocketBase, s *discordgo.Session, m *discordgo.Member) (user User, err error) {
	guild := create_guild(app, s, m.GuildID)
	user = User{}

	err = app.Dao().DB().
		Select("snowflake", "name", "display_name", "guild", "id").
		From("discord_users").
		AndWhere(dbx.In("snowflake", m.User.ID)).
		AndWhere(dbx.In("guild", guild.Id)).
		One(&user)
	return user, err
}

func get_attachment(app *pocketbase.PocketBase, s *discordgo.Session, m string) (attachment Attachment, err error) {
	// return attachment
	err = app.Dao().DB().
		Select("url", "proxy_url", "filename", "id").
		From("discord_attachments").
		AndWhere(dbx.In("id", m)).
		One(&attachment)

	return attachment, err
}

func create_attachments(app *pocketbase.PocketBase, s *discordgo.Session, m *discordgo.Message) (err error) {
	if app == nil || s == nil || m == nil {
		return nil
	}

	collection, err := app.Dao().FindCollectionByNameOrId("discord_attachments")

	if err != nil {
		return err
	}

	for _, attachment := range m.Attachments {
		record := models.NewRecord(collection)

		record.Set("url", attachment.URL)
		record.Set("proxy_url", attachment.ProxyURL)
		record.Set("filename", attachment.Filename)

		if err := app.Dao().SaveRecord(record); err != nil {
			return err
		}
	}
	return nil
}

func create_user(app *pocketbase.PocketBase, s *discordgo.Session, m *discordgo.Member) (err error) {
	if app == nil || s == nil || m == nil {
		return nil
	}
	guild := create_guild(app, s, m.GuildID)

	collection, err := app.Dao().FindCollectionByNameOrId("discord_users")

	if err != nil {
		return err
	}

	record := models.NewRecord(collection)

	if m.User == nil {
		return nil
	}

	record.Set("snowflake", m.User.ID)
	record.Set("name", m.User.Username)
	record.Set("display_name", m.Nick)
	record.Set("guild", guild.Id)

	if err := app.Dao().SaveRecord(record); err != nil {
		return err
	}

	// _, err = app.Dao().DB().
	// 	Insert("discord_users", dbx.Params{
	// 		"snowflake":    m.User.ID,
	// 		"name":         m.User.Username,
	// 		"display_name": m.Nick,
	// 		"guild":        guild.Id,
	// 	}).
	// 	Execute()
	return err
}

func create_message(app *pocketbase.PocketBase, s *discordgo.Session, argMessage *discordgo.Message) (err error) {
	if app == nil || s == nil || argMessage == nil {
		return nil
	}
	channel := create_channel(app, s, argMessage.ChannelID)
	dchannel, err := s.Channel(channel.Snowflake)
	if err != nil {
		return err
	}
	guild := create_guild(app, s, dchannel.GuildID)

	var reference *Message

	if argMessage.MessageReference != nil {
		for reference == nil {
			reference2, _ := get_message(app, s, argMessage.MessageReference.MessageID)

			reference = &reference2

			dmessage, _ := s.ChannelMessage(argMessage.MessageReference.ChannelID, argMessage.MessageReference.MessageID)
			create_message(app, s, dmessage)
		}
	}

	if argMessage.Author == nil {
		return nil
	}

	requestedMessage, err := s.ChannelMessage(argMessage.ChannelID, argMessage.ID)

	if err != nil {
		fmt.Println("b2 utils.go error")
		return err
	}

	s.LogLevel = discordgo.LogDebug
	// s.Debug = true

	// apipoint := discordgo.EndpointGuildMember(requestedMessage.GuildID, requestedMessage.Author.ID)

	// body, err := s.Client.Get(apipoint)

	fmt.Println(requestedMessage.ChannelID, requestedMessage.GuildID, dchannel.GuildID)

	member, err := s.GuildMember(dchannel.GuildID, requestedMessage.Author.ID)
	if err != nil {
		fmt.Println("b3 utils.go error")
		return err
	}

	create_user(app, s, member)

	author, _ := get_user(app, s, member)

	collection, err := app.Dao().FindCollectionByNameOrId("discord_messages")

	if err != nil {
		return err
	}

	field_collection, err := app.Dao().FindCollectionByNameOrId("discord_embed_fields")
	if err != nil {
		return err
	}
	embed_collection, err := app.Dao().FindCollectionByNameOrId("discord_embeds")

	if err != nil {
		return err
	}

	var embeds []string

	for _, embed := range argMessage.Embeds {
		var fields []string
		for _, field := range embed.Fields {

			record := models.NewRecord(field_collection)

			record.Set("name", field.Name)
			record.Set("value", field.Value)
			record.Set("inline", field.Inline)

			if err := app.Dao().SaveRecord(record); err != nil {
				return err
			}

			// _, err = app.Dao().DB().
			// 	Insert("discord_embed_fields", dbx.Params{
			// 		"name":   field.Name,
			// 		"value":  field.Value,
			// 		"inline": field.Inline,
			// 	}).
			// 	Execute()

			if err != nil {
				return err
			}
			field2 := EmbedField{}

			err = app.Dao().DB().
				Select("id").
				From("discord_embed_fields").
				AndWhere(dbx.In("name", field.Name)).
				AndWhere(dbx.In("value", field.Value)).
				AndWhere(dbx.In("inline", field.Inline)).
				One(&field2)

			fields = append(fields, field2.Id)
			if err != nil {
				continue
			}
		}
		record := models.NewRecord(embed_collection)

		record.Set("title", embed.Title)
		record.Set("type", embed.Type)
		record.Set("url", embed.URL)
		record.Set("timestamp", embed.Timestamp)
		record.Set("color", embed.Color)
		if embed.Footer != nil {
			record.Set("footer_text", embed.Footer.Text)
			record.Set("footer_iconurl", embed.Footer.IconURL)
		}
		if embed.Video != nil {
			record.Set("video", embed.Video.URL)
		}
		if embed.Provider != nil {
			record.Set("provider", embed.Provider.Name)
		}
		record.Set("fields", fields)
		if embed.Image != nil {
			record.Set("image", embed.Image.URL)
		}
		if embed.Thumbnail != nil {
			record.Set("thumbnail", embed.Thumbnail.URL)
		}

		record.MarkAsNew()

		if err := app.Dao().SaveRecord(record); err != nil {
			continue
		}

		// _, err = app.Dao().DB().
		// 	Insert("discord_embeds", dbx.Params{
		// 		"title":          embed.Title,
		// 		"type":           embed.Type,
		// 		"url":            embed.URL,
		// 		"timestamp":      embed.Timestamp,
		// 		"color":          embed.Color,
		// 		"footer_text":    embed.Footer.Text,
		// 		"footer_iconurl": embed.Footer.IconURL,
		// 		"video":          embed.Video,
		// 		"provider":       embed.Provider.Name,
		// 		"fields":         fields,
		// 		"image":          embed.Image.URL,
		// 		"thumbnail":      embed.Thumbnail.URL,
		// 	}).
		// 	Execute()

		if err != nil {
			return err
		}

		embed2 := Embed{}

		err = app.Dao().DB().
			Select("id").
			From("discord_embeds").
			AndWhere(dbx.In("title", embed.Title)).
			AndWhere(dbx.In("type", embed.Type)).
			AndWhere(dbx.In("url", embed.URL)).
			AndWhere(dbx.In("timestamp", embed.Timestamp)).
			One(&embed2)

		if err != nil {
			return err
		}
		fmt.Println("embed made!")
	}

	err = create_attachments(app, s, argMessage)

	attachments := []string{}

	for _, attachment := range argMessage.Attachments {
		attachment2, err := get_attachment(app, s, attachment.ID)
		if err != nil {
			return err
		}
		attachments = append(attachments, attachment2.Id)
	}

	record := models.NewRecord(collection)
	record.Set("snowflake", argMessage.ID)
	record.Set("channel", channel.Id)
	record.Set("guild", guild.Id)
	record.Set("content", argMessage.Content)
	record.Set("author", author.Id)
	if reference != nil {
		record.Set("reference", reference.Id)
	}
	record.Set("embeds", embeds)
	if argMessage.EditedTimestamp != nil {
		record.Set("edited", argMessage.EditedTimestamp)
	}
	record.Set("attachments", attachments)

	record.MarkAsNew()

	if err := app.Dao().SaveRecord(record); err != nil {
		return err
	}

	// _, err = app.Dao().DB().
	// 	Insert("discord_messages", dbx.Params{
	// 		"snowflake": m.ID,
	// 		"channel":   channel.Id,
	// 		"guild":     guild.Id,
	// 		"content":   m.Content,
	// 		"author":    author.Id,
	// 		"reference": reference,
	// 		"embeds":    embeds,
	// 	}).
	// 	Execute()

	if err != nil {
		return err
	}
	return nil
}
