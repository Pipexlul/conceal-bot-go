package mappers

import (
	"github.com/bwmarrin/discordgo"

	"github.com/pipexlul/conceal-bot-go/internal/models"
	utils "github.com/pipexlul/conceal-bot-go/internal/utilities"
)

var (
	botEventTypeToDiscordGoEventType = map[models.BotEventType]utils.DiscordGoHandlerWrapper{
		models.EventType_Ready: func(bot utils.ConcealBot, handler utils.ConcealBotHandler) interface{} {
			return func(session *discordgo.Session, eventData *discordgo.Ready) {
				handler(bot, session, eventData)
			}
		},
	}
)

func GetDiscordGoHandlerWrapperForBotEventType(eventType models.BotEventType) utils.DiscordGoHandlerWrapper {
	if wrapper, ok := botEventTypeToDiscordGoEventType[eventType]; ok {
		return wrapper
	}
	return nil
}
