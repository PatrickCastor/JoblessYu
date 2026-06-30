package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func buildCompanyLogoURL(jobURL string) string {
	if jobURL == "" {
		return ""
	}

	parsedURL, err := url.Parse(jobURL)
	if err != nil || parsedURL.Host == "" {
		return ""
	}

	host := strings.ToLower(parsedURL.Host)
	host = strings.TrimPrefix(host, "www.")

	return fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=64", host)
}

func buildJobsEmbed(ctx context.Context, dbURL string) (*discordgo.MessageEmbed, error) {
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return nil, err
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
		return nil, err
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
	var logoURL string

	for rows.Next() {
		var title, company, location, url string

		if err := rows.Scan(&title, &company, &location, &url); err != nil {
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

		if logoURL == "" {
			logoURL = buildCompanyLogoURL(url)
		}

		jobCount++
	}

	if jobCount == 0 {
		return nil, fmt.Errorf("no jobs found")
	}

	if logoURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: logoURL}
	}

	return embed, nil
}

func main() {
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
		if m.Author.Bot {
			return
		}

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

		if m.Content == "!jobs" {
			s.ChannelMessageSend(m.ChannelID, "🔍 Checking Neon DB for latest jobs...")

			ctx := context.Background()
			embed, err := buildJobsEmbed(ctx, dbURL)
			if err != nil {
				if err.Error() == "no jobs found" {
					s.ChannelMessageSend(m.ChannelID, "⚠️ No jobs found in the database.")
				} else {
					s.ChannelMessageSend(m.ChannelID, "❌ Failed to fetch jobs.")
					log.Println("Jobs Build Error:", err)
				}
				return
			}

			_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
			if err != nil {
				log.Println("Send Embed Error:", err)
				s.ChannelMessageSend(m.ChannelID, "❌ Failed to send embed.")
			}
		}
	})

	disbot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		cmd := i.ApplicationCommandData().Name
		switch cmd {
		case "testembed":
			embed := &discordgo.MessageEmbed{
				Title:       "✅ Embed Test",
				Description: "If you can see this, embeds are working correctly.",
				Color:       0x57F287,
			}

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
			})
			if err != nil {
				log.Println("Interaction Respond Error:", err)
			}
		case "jobs":
			log.Println("Received jobs interaction")
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "🔍 Fetching the latest jobs...",
				},
			})
			if err != nil {
				log.Println("Interaction Respond Error:", err)
				return
			}

			go func() {
				ctx := context.Background()
				embed, err := buildJobsEmbed(ctx, dbURL)
				if err != nil {
					msg := "⚠️ No jobs found in the database."
					if err.Error() != "no jobs found" {
						msg = "❌ Failed to fetch jobs."
						log.Println("Jobs Build Error:", err)
					}

					_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{Content: msg})
					if err != nil {
						log.Println("FollowupMessageCreate Error:", err)
					}
					return
				}

				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{Embeds: []*discordgo.MessageEmbed{embed}})
				if err != nil {
					log.Println("FollowupMessageCreate Error:", err)
				}
			}()
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
