package main

import (
	"fmt"       // For printing to your Linux terminal
	"os"        // To talk to your EndeavourOS system
	"os/signal" // To handle "Ctrl+C" stops
	"syscall"   // For low-level system signals

	"github.com/bwmarrin/discordgo" // The library you just 'go get'-ted
)

func submain() {
	disbot, err := discordgo.New("Bot MTUwMzYxOTU5NDAyNjM1MjY5MA.GZLx3M.__JFfghZUWTLKIFVqv4ObsGqayW232SmAzQR28")
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	//Permission set:
	disbot.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages
	//Connection status:
	//Error
	err = disbot.Open()
	if err != nil {
		fmt.Println("Error opening connection", err)
		return
	}
	//Success
	fmt.Println("Bot is now running. Press CTRL-C to exit.")

	//Ctrl + C signal handling:
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	//Ctrl + C signal received, cleanly close down the Discord session:
	fmt.Println("Shutting down")
	disbot.Close()
}
