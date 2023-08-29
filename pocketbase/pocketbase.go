package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	linuxproc "github.com/c9s/goprocinfo/linux"
	"github.com/goccy/go-json"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Config struct {
	Info       string `json:"_info"`
	PBInstance struct {
		URL           string `json:"url"`
		AdminEmail    string `json:"admin_email"`
		AdminPassword string `json:"admin_password"`
	} `json:"pb_instance"`
	DiscordToken      string `json:"discord_token"`
	RerunMessages     string `json:"rerun_messages"`
	AttachmentChannel string `json:"attachment_channel"`
}

var config Config

func main() {

	file, err := os.Open("../config_secret.json")

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	defer func() {
		if err := recover(); err != nil {
			// Handle the panic gracefully
			fmt.Println("Panic occurred:", err)
		}
	}()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	app := pocketbase.New()

	discord, err := discordgo.New("Bot " + config.DiscordToken)

	starttime := time.Now()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/api/botstats", func(c echo.Context) error {
			stat, err := linuxproc.ReadStat("/proc/stat")

			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, map[string]string{"uptime": starttime.String(), "cpu": fmt.Sprint(stat.CPUStats[0].User + stat.CPUStats[0].System), "compiler_go_version": runtime.Version(), "go_version": runtime.Version(), "num_goroutine": fmt.Sprint(runtime.NumGoroutine())})
		})

		return nil
	})

	discord.Identify.Intents = discordgo.IntentsAll

	if err != nil {
		log.Fatal(err)
	}

	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		dispatch_message(app, s, m)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		dispatch_message_update(app, s, m)
	})

	if config.RerunMessages == "true" {
		discord.AddHandlerOnce(func(s *discordgo.Session, m *discordgo.Ready) {
			fmt.Println("Bot is up!")
			fmt.Println(len(m.Guilds))

			wg := sync.WaitGroup{}

			for a, guild := range m.Guilds {
				guild3, err := s.Guild(guild.ID)
				fmt.Println(guild3.Name)
				fmt.Println(guild3.ID)
				if err != nil {
					continue
				}
				create_guild(app, s, guild3.ID)

				fmt.Println(a)
				channels, _ := s.GuildChannels(guild3.ID)
				for _, channel := range channels {
					fmt.Println(channel.Name)
					wg.Add(1)
					go func(app *pocketbase.PocketBase, s *discordgo.Session, channel *discordgo.Channel) {
						create_channel(app, s, channel.ID)
						wg.Done()
					}(app, s, channel)
					wg.Add(1)
					go func(app *pocketbase.PocketBase, s *discordgo.Session, channel *discordgo.Channel) {
						log_channel(app, s, channel.ID)
						wg.Done()
					}(app, s, channel)
				}
				for _, member := range guild3.Members {
					fmt.Println(create_user(app, s, member))
				}
			}
			wg.Wait()
		})
	}

	if err != nil {
		log.Fatal(err)
	}

	discord.LogLevel = discordgo.LogInformational

	// connect to discord without blocking
	go discord.Open()

	if err := app.Start(); err != nil {
		log.Fatalln(err)
	}
}
