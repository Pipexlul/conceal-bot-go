package utilities

import (
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/pipexlul/conceal-bot-go/internal/models"
	discordstatus "github.com/pipexlul/conceal-bot-go/internal/pkg/discord-status"
)

type ConcealBot interface {
	ConnectMongo()
	DisconnectMongo()
	RegisterHandler(handler ConcealBotHandler, handlerType models.BotEventType) func()
	InjectBotForHandler(handler ConcealBotHandler, handlerType models.BotEventType) interface{}

	Client() *discordgo.Session

	GetMongoClient() *mongo.Client
	GetStatusHelper() *discordstatus.Helper
	GetRandGen() *rand.Rand
}

type ConcealBotHandler func(ConcealBot, *discordgo.Session, interface{})

type DiscordGoHandlerWrapper func(ConcealBot, ConcealBotHandler) interface{}

type BotEventTypeMap map[models.BotEventType]DiscordGoHandlerWrapper
