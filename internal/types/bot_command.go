package types

import (
	"context"
	utils "github.com/pipexlul/conceal-bot-go/internal/utilities"

	"github.com/bwmarrin/discordgo"
)

type BotCommand interface {
	Register(ctx context.Context, concealBot utils.ConcealBot) error
	// Execute(ctx context.Context) error
	GetCommandName() string
	GetOptions() []*discordgo.ApplicationCommandOption
}
