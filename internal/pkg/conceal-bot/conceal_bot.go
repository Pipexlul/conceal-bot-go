package conceal_bot

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/pipexlul/conceal-bot-go/internal/mappers"
	"github.com/pipexlul/conceal-bot-go/internal/models"
	discordstatus "github.com/pipexlul/conceal-bot-go/internal/pkg/discord-status"
	utils "github.com/pipexlul/conceal-bot-go/internal/utilities"
)

type ConcealBot struct {
	DiscordClient *discordgo.Session
	StatusHelper  *discordstatus.Helper
	RandGen       *rand.Rand
	MongoClient   *mongo.Client
}

func NewConcealBot(options ...ConcealBotOption) *ConcealBot {
	bot := &ConcealBot{}
	for _, option := range options {
		option(bot)
	}

	return bot
}

func (bot *ConcealBot) Client() *discordgo.Session {
	return bot.DiscordClient
}

func (bot *ConcealBot) GetMongoClient() *mongo.Client {
	return bot.MongoClient
}

func (bot *ConcealBot) GetStatusHelper() *discordstatus.Helper {
	return bot.StatusHelper
}

func (bot *ConcealBot) GetRandGen() *rand.Rand {
	return bot.RandGen
}

func (bot *ConcealBot) ConnectMongo() {
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		log.Fatal("Missing MONGO_URL environment variable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB")
	bot.MongoClient = client
}

func (bot *ConcealBot) DisconnectMongo() {
	if bot.MongoClient != nil {
		if err := bot.MongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}

		log.Println("Disconnected from MongoDB")
	}
}

func (bot *ConcealBot) RegisterHandler(handler utils.ConcealBotHandler, handlerType models.BotEventType) func() {
	return bot.DiscordClient.AddHandler(bot.InjectBotForHandler(handler, handlerType))
}

func (bot *ConcealBot) InjectBotForHandler(handler utils.ConcealBotHandler, handlerType models.BotEventType) interface{} {
	handlerWrapper := mappers.GetDiscordGoHandlerWrapperForBotEventType(handlerType)

	if handlerWrapper == nil {
		log.Fatalf("Unknown handler type: %v", handlerType)
	}

	return handlerWrapper(bot, handler)
}

func (bot *ConcealBot) SetStatusHelper(statusHelper *discordstatus.Helper) {
	bot.StatusHelper = statusHelper
}

func (bot *ConcealBot) Validate() {
	if bot.DiscordClient == nil {
		log.Fatalln("Missing Discord client")
	}
	if bot.StatusHelper == nil {
		log.Fatalln("Missing Discord status helper")
	}
	if bot.RandGen == nil {
		log.Fatalln("Missing random number generator")
	}
	if bot.MongoClient == nil {
		log.Fatalln("Missing MongoDB client")
	}
}

type ConcealBotOption func(bot *ConcealBot)

func WithRandGen(randGen *rand.Rand) ConcealBotOption {
	return func(bot *ConcealBot) {
		bot.RandGen = randGen
	}
}
