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
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found, using system env variables")
	}

	token := "Bot " + os.Getenv("DISCORD_BOT_TOKEN")
	dbURL := os.Getenv("DATABASE_URL")

	disbot, err := discordgo.New(token)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}

	disbot.Identify.Intents =
		discordgo.IntentsGuildMessages |
			discordgo.IntentsDirectMessages |
			discordgo.IntentsGuilds |
			discordgo.IntentsMessageContent

	disbot.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		// Ignore bot messages
		if m.Author.Bot {
			return
		}

		// Test embed command
		if m.Content == "!testembed" {
			embed := &discordgo.MessageEmbed{
				Title:       "✅ Embed Test",
				Description: "If you can see this, embeds are working correctly.",
				Color:       0x57F287,
			}

			_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
			if err != nil {
				log.Println("Embed Error:", err)
			}
			return
		}

		// Jobs command
		if m.Content == "!jobs" {

			s.ChannelMessageSend(
				m.ChannelID,
				"🔍 Checking Neon DB for latest jobs...",
			)

			ctx := context.Background()

			conn, err := pgx.Connect(ctx, dbURL)
			if err != nil {
				s.ChannelMessageSend(
					m.ChannelID,
					"❌ Failed to connect to Neon DB.",
				)
				log.Println("DB Connect Error:", err)
				return
			}
			defer conn.Close(ctx)

			rows, err := conn.Query(
				ctx,
				`SELECT title, company, location, job_url
				 FROM jobs
				 ORDER BY fetched_at DESC
				 LIMIT 10`,
			)
			if err != nil {
				s.ChannelMessageSend(
					m.ChannelID,
					"❌ Failed to query jobs.",
				)
				log.Println("Query Error:", err)
				return
			}
			defer rows.Close()

			embed := &discordgo.MessageEmbed{
				Title:       "🇻🇳 Latest Jobs",
				Description: "Newest jobs found in the database.",
				Color:       0x5865F2,
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Powered by JobSpy + Neon",
				},
			}

			jobCount := 0

			for rows.Next() {
				var title, company, location, url string

				if err := rows.Scan(
					&title,
					&company,
					&location,
					&url,
				); err != nil {
					continue
				}

				embed.Fields = append(
					embed.Fields,
					&discordgo.MessageEmbedField{
						Name: title,
						Value: fmt.Sprintf(
							"🏢 %s\n📍 %s\n🔗 %s",
							company,
							location,
							url,
						),
						Inline: false,
					},
				)

				jobCount++
			}

			if jobCount == 0 {
				s.ChannelMessageSend(
					m.ChannelID,
					"⚠️ No jobs found in the database.",
				)
				return
			}

			_, err = s.ChannelMessageSendEmbed(
				m.ChannelID,
				embed,
			)

			if err != nil {
				log.Println("Send Embed Error:", err)
				s.ChannelMessageSend(
					m.ChannelID,
					"❌ Failed to send embed.",
				)
			}
		}
	})

	if err := disbot.Open(); err != nil {
		log.Fatal("Error opening connection:", err)
	}

	fmt.Println("JoblessYu Vessel is now running. Press CTRL+C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("Shutting down...")
	disbot.Close()
}
