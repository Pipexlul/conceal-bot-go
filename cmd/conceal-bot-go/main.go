package main

import (
	"context"
	"log"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/pipexlul/conceal-bot-go/internal/commands"
	"github.com/pipexlul/conceal-bot-go/internal/mappers"
	"github.com/pipexlul/conceal-bot-go/internal/pkg/env"
)

const funnyStatus = "Middle name?!?!"

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
	botToken := env.GetBotToken()
	if botToken.Token == "" {
		log.Fatal("Missing all BOT TOKEN environment variables, at least one is required")
	}

	log.Printf("Starting bot in %s mode", mappers.MapTokenTypeToString(botToken.TokenType))

	dg, err := discordgo.New("Bot " + botToken.Token)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

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

	registererCtx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()

	// TODO: Use a more centralized command registerer
	cmd := &commands.TimeDifferenceCmd{}
	registerCmdErr := cmd.Register(registererCtx, dg)
	if registerCmdErr != nil {
		log.Fatalf("Failed to register command: %v", cmd.GetCommandName())
	}

	log.Print("All commands registered!")

	log.Println("Bot is now running. Press CTRL-C to exit.")
	select {}
}

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	if err := s.UpdateGameStatus(0, funnyStatus); err != nil {
		log.Printf("Error updating game status at ready: %v", err)
	}
	log.Printf("Ready! Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
}
