package utilities

import (
	"github.com/bwmarrin/discordgo"

	"github.com/pipexlul/conceal-bot-go/internal/models"
	discordstatus "github.com/pipexlul/conceal-bot-go/internal/pkg/discord-status"
)

type ConcealBot interface {
	ConnectMongo()
	DisconnectMongo()
	RegisterHandler(handler ConcealBotHandler, handlerType models.BotEventType) func()
	InjectBotForHandler(handler ConcealBotHandler, handlerType models.BotEventType) interface{}

	GetStatusHelper() *discordstatus.Helper
}

type ConcealBotHandler func(ConcealBot, *discordgo.Session, interface{})

type DiscordGoHandlerWrapper func(ConcealBot, ConcealBotHandler) interface{}

type BotEventTypeMap map[models.BotEventType]DiscordGoHandlerWrapper
