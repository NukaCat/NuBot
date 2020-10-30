package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	session, err := discordgo.New("Bot NzcxNDAxMTA5OTE4MDU2NDU5.X5rlRA.m85A7K6HBh0drFo4IlUEnVpaP9s")
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	session.AddHandler(messageCreate)

	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)

	// Open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	session.Close()
}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {

	if message.Author.ID == session.State.User.ID {
		return
	}

	str := message.Content
	if strings.HasPrefix(str, "connect") {
		words := strings.Split(str, " ")
		if len(words) < 2 {
			fmt.Println("invalid command")
			return
		}
		channelID := words[1]
		guild, err := session.Guild(message.GuildID)
		if err != nil {
			fmt.Println("guild is is invalid")
			return
		}
		session.ChannelVoiceJoin(guild.ID, channelID, false, false)
	}

	if strings.HasPrefix(str, "play") {
		words := strings.Split(str, " ")
		if len(words) < 2 {
			fmt.Println("invalid command")
			return
		}
		channel, exist := session.VoiceConnections[message.GuildID]
		if !exist {
			fmt.Println("Can't play sound without channel")
			return
		}
		playSound(channel, words[1]);
	}
}

func loadSound(name string) ([][]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return nil, err
	}

	var opuslen int16
	ret := make([][]byte, 0)

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return nil, err
			}
			return ret, nil
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return nil, err
		}

		// Read encoded pcm from dca file.
		buf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &buf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return nil, err
		}

		// Append encoded pcm data to the buffer.
		ret = append(ret, buf)
	}
}

// playSound plays the current buffer to the provided channel.
func playSound(channel *discordgo.VoiceConnection, sound string) (err error) {
	buffer, err := loadSound(sound)
	if err != nil {
		return err
	}

	channel.Speaking(true)
	for _, buff := range buffer {
		channel.OpusSend <- buff
	}
	channel.Speaking(false)
	return nil
}
