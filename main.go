package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	Token         string
	ChannelTarget string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot token")
	flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages

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
	userMessageSplit := strings.Split(m.Content, " ")
	ChannelTarget = strings.TrimPrefix(strings.TrimSuffix(userMessageSplit[1], ">"), "<#")
	s.ChannelMessageSend(m.ChannelID, "Set target notes channel to "+userMessageSplit[1])
}

func WriteNoteToChannel(s *discordgo.Session, m *discordgo.MessageCreate) {
	isReference := false
	var ReferencedMessageLink string
	if m.Message.ReferencedMessage != nil {
		isReference = true
	}
	if isReference {
		ReferenceID := m.Message.ReferencedMessage.ID
		GuildID := m.GuildID
		ChannelID := m.ChannelID
		ReferencedMessageLink = "https://discordapp.com/channels/" + GuildID + "/" + ChannelID + "/" + ReferenceID
	}

	messageContents := strings.TrimPrefix(m.Content, "!dsnote")
	s.ChannelMessageSend(ChannelTarget, ReferencedMessageLink+" "+messageContents)
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

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")

	}

	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}
