package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load the .env file (now that you have godotenv installed)
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, using system env variables")
	}

	// Use your reset token here once you have it!
	token := "Bot " + os.Getenv("DISCORD_BOT_TOKEN")
	dbURL := os.Getenv("DATABASE_URL")

	disbot, err := discordgo.New(token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// 2. Add the Handler: This tells the bot what to do when a message arrives
	disbot.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore messages from the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}

		// Trigger command: !jobs
		if m.Content == "!jobs" {
			s.ChannelMessageSend(m.ChannelID, "Checking Neon DB for IT Support roles in Vietnam... =w=")

			ctx := context.Background()
			conn, err := pgx.Connect(ctx, dbURL)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "❌ Failed to connect to Neon DB.")
				fmt.Println("DB Connect Error:", err)
				return
			}
			defer conn.Close(ctx)

			// Query the latest 10 jobs scraped by JobSpy
			rows, _ := conn.Query(ctx, "SELECT title, company, location, job_url FROM jobs ORDER BY fetched_at DESC LIMIT 10")

			for rows.Next() {
				var title, company, loc, url string
				err := rows.Scan(&title, &company, &loc, &url)
				if err != nil {
					continue
				}

				// Post to Discord
				msg := fmt.Sprintf("📌 **%s**\n🏢 %s | 📍 %s\n🔗 <%s>", title, company, loc, url)
				s.ChannelMessageSend(m.ChannelID, msg)
			}
		}
	})
	//Permission set:
	disbot.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	err = disbot.Open()
	if err != nil {
		fmt.Println("Error opening connection", err)
		return
	}

	fmt.Println("JoblessYu Vessel is now running. Press CTRL-C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("Shutting down")
	disbot.Close()
}
