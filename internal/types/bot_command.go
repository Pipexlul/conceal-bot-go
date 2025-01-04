package types

import (
	"context"

	"github.com/bwmarrin/discordgo"

	utils "github.com/pipexlul/conceal-bot-go/internal/utilities"
)

type BotCommand interface {
	Register(ctx context.Context, concealBot utils.ConcealBot) error
	// Execute(ctx context.Context) error
	GetCommandName() string
	GetOptions() []*discordgo.ApplicationCommandOption
}
