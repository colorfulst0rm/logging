package main

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pocketbase/pocketbase"
)

func GetAllMessagesInChannel(app *pocketbase.PocketBase, s *discordgo.Session, channelID string) (err error) {

	wg := sync.WaitGroup{}

	// Initialize variables for pagination
	var beforeID, afterID, aroundID string
	var numMessages int

	for {
		// Retrieve the next batch of messages
		messages, err := s.ChannelMessages(channelID, numMessages, beforeID, afterID, aroundID)
		if err != nil {
			// Handle error, break the loop, or take other appropriate action
			break
		}

		// If no more messages are returned, break the loop
		if len(messages) == 0 {
			break
		}

		// Update pagination variables for the next iteration
		beforeID = messages[len(messages)-1].ID
		afterID = ""
		aroundID = ""

		for _, message := range messages {
			wg.Add(1)
			go func(message *discordgo.Message, s *discordgo.Session) {

				fmt.Println(create_message(app, s, message))
				wg.Done()
			}(message, s)
		}

		// Check if the retrieved message count is less than the requested count
		// This means you've fetched all available messages
		if len(messages) < numMessages {
			break
		}
	}
	wg.Wait()
	return nil
}

func log_channel(app *pocketbase.PocketBase, s *discordgo.Session, channelID string) {
	GetAllMessagesInChannel(app, s, channelID)
}
