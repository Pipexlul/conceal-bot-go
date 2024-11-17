package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func connectMongo() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("Missing MONGO_URI environment variable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB")
	mongoClient = client
}

func disconnectMongo() {
	if mongoClient != nil {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}
}

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Missing BOT_TOKEN environment variable")
	}

	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand && i.ApplicationCommandData().Name == "cctime" {
			handleConvertTime(s, i)
		}
	})

	dg.AddHandler(onReady)

	if err = dg.Open(); err != nil {
		log.Fatalf("Failed to open Discord session: %v", err)
	}
	log.Println("Discord session opened :)")

	defer func() {
		closeErr := dg.Close()
		if closeErr != nil {
			log.Fatalf("Failed to close Discord session: %v", closeErr)
		}
		log.Println("Discord session closed")
	}()

	_, err = dg.ApplicationCommandCreate(dg.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        "cctime",
		Description: "Convert time between New Jersey/Philadelphia and Chile",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "time",
				Description: "Time to convert",
				Required:    true,
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create command: %v", err)
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")
	select {}
}

func handleConvertTime(s *discordgo.Session, i *discordgo.InteractionCreate) {
	const timeFormat = "15:04"

	timeStr := i.ApplicationCommandData().Options[0].StringValue()
	parsedTime, err := time.Parse(timeFormat, timeStr)
	if err != nil {
		replyWithMessage(s, i, "Invalid time format. Please use HH:MM. (24-hour format)")
		return
	}

	timezonesMap := map[string]string{
		"New Jersey/Philadelphia": "America/New_York",
		"Chile":                   "America/Santiago",
	}

	var results []string
	for locationName, timezone := range timezonesMap {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			log.Printf("Error loading location %s: %v", timezone, err)
			continue
		}

		now := time.Now()
		timeInLocation := time.Date(now.Year(), now.Month(), now.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, loc)
		results = append(results, fmt.Sprintf("If %s is the time in %s, then:", timeStr, locationName))

		for targetName, targetTimezone := range timezonesMap {
			if targetName == locationName {
				continue
			}
			targetLoc, err := time.LoadLocation(targetTimezone)
			if err != nil {
				log.Printf("Error loading location %s: %v", targetTimezone, err)
				continue
			}
			targetTime := timeInLocation.In(targetLoc)
			results = append(results, fmt.Sprintf("- %s would be the time in %s", targetTime.Format(timeFormat), targetName))
		}
		results = append(results, "")
	}
	replyWithMessage(s, i, strings.Join(results, "\n"))
}

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	if err := s.UpdateGameStatus(0, "Concealing lol"); err != nil {
		log.Printf("Error updating game status at ready: %v", err)
	}
	log.Printf("Ready! Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
}

func replyWithMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
	if err != nil {
		log.Printf("Error sending response: %v", err)
	}
}
