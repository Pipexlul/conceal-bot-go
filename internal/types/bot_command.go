package types

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type BotCommandHandler func(dgSession *discordgo.Session, i *discordgo.InteractionCreate)

type BotCommand interface {
	Register(ctx context.Context, dgSession *discordgo.Session) error
	// Execute(ctx context.Context) error
	GetCommandName() string
	GetOptions() []*discordgo.ApplicationCommandOption
}
