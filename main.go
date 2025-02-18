package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	config         DSConfig
	configLocation string
)

func saveConfig() {
	if configLocation != "TOKEN" {
		data, err := json.Marshal(config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.WriteFile(configLocation, data, 0644)
	}
}

func init() {
	flag.StringVar(&config.BotToken, "t", "", "Bot token")
	flag.Parse()

	UserHome, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ConfigHome, err := os.UserConfigDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if config.BotToken == "" {
		ConfigData, err := os.ReadFile(path.Join(UserHome, ".dsnote.json"))
		if err != nil {
			ConfigData, err = os.ReadFile(path.Join(ConfigHome, "dsnote", "config.json"))
			if err != nil {
				fmt.Println("Unable to obtain token.")
				fmt.Println("Either you forgot to supply a token with the -t option or no configuration file could be found.")
				os.Exit(1)
			} else {
				configLocation = path.Join(ConfigHome, "dsnote", "config.json")
			}
		} else {
			configLocation = path.Join(UserHome, ".dsnote.json")
		}
		err = json.Unmarshal(ConfigData, &config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		configLocation = "TOKEN"
	}
}

func main() {
	defer saveConfig()
	dg, err := discordgo.New("Bot " + config.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages + discordgo.IntentsGuilds

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func SetTargetChannel(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild := config.GetServer(m.GuildID)
	userMessageContent := strings.TrimSpace(strings.TrimPrefix(m.Content, guild.Prefix+"set_target_channel"))
	if len(userMessageContent) > 0 {
		targetChannel := strings.TrimPrefix(strings.TrimSuffix(userMessageContent, ">"), "<#")
		guild.SetTargetChannel(targetChannel)
		s.ChannelMessageSend(m.ChannelID, "Set target notes channel to <#"+targetChannel+">")
		s.ChannelMessageDelete(m.ChannelID, m.Message.ID)
	} else {
		s.ChannelMessageSend(m.ChannelID, "Couldn't set target channel, did you supply one?")
	}
}

func DeleteMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Message.ReferencedMessage != nil {
		s.ChannelMessageDelete(m.Message.ReferencedMessage.ChannelID, m.Message.ReferencedMessage.ID)
		s.ChannelMessageDelete(m.Message.ChannelID, m.Message.ID)
	}
}

func EditMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild := config.GetServer(m.GuildID)
	messageContent := strings.TrimSpace(strings.TrimPrefix(m.Content, guild.Prefix+"dsedit"))
	if len(messageContent) > 0 {
		if m.Message.ReferencedMessage != nil {
			_, err := s.ChannelMessageEdit(
				m.ReferencedMessage.ChannelID,
				m.ReferencedMessage.ID,
				messageContent)
			if err != nil {
				userChannel, err := s.UserChannelCreate(m.Author.ID)
				if err != nil {
					fmt.Println("Oops unable to send message.")
					return
				}
				s.ChannelMessageSend(userChannel.ID, "Unable to edit requested message.")
			}
		}
		s.ChannelMessageDelete(m.Message.ChannelID, m.Message.ID)
	}
}



func WriteNoteToChannel(s *discordgo.Session, m *discordgo.MessageCreate) {
	guildID := config.GetServer(m.GuildID)

	var referencedMessageLink string
	if m.Message.ReferencedMessage != nil {
		referenceID := m.Message.ReferencedMessage.ID
		channelID := m.ChannelID
		referencedMessageLink = "https://discordapp.com/channels/" + guildID.GuildID + "/" + channelID + "/" + referenceID
	}

	messageContents := strings.TrimSpace(strings.TrimPrefix(m.Content, guildID.Prefix+"dsnote"))
	if len(messageContents) > 0 {
		s.ChannelMessageSend(guildID.TargetChannel, messageContents+" "+referencedMessageLink)
		s.ChannelMessageDelete(m.ChannelID, m.Message.ID)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(strings.ToLower(m.Content), "!set_target_channel") {
		SetTargetChannel(s, m)
	}

	if strings.HasPrefix(strings.ToLower(m.Content), "!dsnote") {
		WriteNoteToChannel(s, m)
	}

	if strings.HasPrefix(strings.ToLower(m.Content), "!dsedit") {
		EditMessage(s, m)
	}

	if strings.HasPrefix(strings.ToLower(m.Content), "!dskill") {
		DeleteMessage(s, m)
	}

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")

	}

	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}
