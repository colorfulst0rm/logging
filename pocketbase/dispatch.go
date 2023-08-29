package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

func dispatch_message(app *pocketbase.PocketBase, s *discordgo.Session, m *discordgo.MessageCreate) {
	create_guild(app, s, m.GuildID)
	create_channel(app, s, m.ChannelID)

	create_user(app, s, m.Member)
	create_message(app, s, m.Message)

	// create author

}

func dispatch_message_update(app *pocketbase.PocketBase, s *discordgo.Session, m *discordgo.MessageUpdate) {
	create_guild(app, s, m.GuildID)
	create_channel(app, s, m.ChannelID)

	create_user(app, s, m.Member)
	create_message_log(app, s, m.Message)
}

func create_message_log(app *pocketbase.PocketBase, s *discordgo.Session, message *discordgo.Message) {
	old_message, err := get_message(app, s, message.ID)
	if err != nil {
		return
	}
	archive_attachments(app, s, old_message)
	collection, err := app.Dao().FindCollectionByNameOrId("discord_message_logs")
	if err != nil {
		return
	}
	record := models.NewRecord(collection)
	record.Set("message", old_message.Id)
	record.Set("old_content", old_message.Content)
	record.Set("new_content", message.Content)

	if err := app.Dao().SaveRecord(record); err != nil {
		return
	}

	record, err = app.Dao().FindRecordById("discord_messages", message.ID)

	if err != nil {
		return
	}

	record.Set("content", message.Content)

	if err := app.Dao().SaveRecord(record); err != nil {
		return
	}

}

func archive_attachments(app *pocketbase.PocketBase, s *discordgo.Session, old_message Message) (err error) {
	old_message_attachments := old_message.Attachments

	var new_attachments []string
	var proxy_attachments []string

	for _, attachment := range old_message_attachments {
		// download attachment into ram
		// send attachment to channel

		attachment2, err := get_attachment(app, s, attachment)
		if err != nil {
			return err
		}

		resp, err := s.Client.Get(attachment)

		if err != nil {
			return err
		}

		defer resp.Body.Close()

		channel, err := s.Channel(config.AttachmentChannel)
		if err != nil {
			return err
		}
		msg, err := s.ChannelFileSendWithMessage(channel.ID, "Attachment deleted: ", attachment2.Filename, resp.Body)
		if err != nil {
			return err
		}

		new_attachments = append(new_attachments, msg.Attachments[0].URL)
		proxy_attachments = append(proxy_attachments, attachment2.ProxyUrl)
	}
	modify_attachment(app, s, new_attachments, proxy_attachments, old_message_attachments)
	return nil
}

func modify_attachment(app *pocketbase.PocketBase, s *discordgo.Session, new_attachments []string, proxy_links []string, old_message_attachments []string) {
	// update attachment in db

	for idx, attachment := range old_message_attachments {
		record, err := app.Dao().FindRecordById("attachments", attachment)
		if err != nil {
			return
		}
		record.Set("url", new_attachments[idx])
		record.Set("proxy_url", proxy_links[idx])
	}
}
