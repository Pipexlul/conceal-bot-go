package main

import (
	"context"
	"log"
	"math/rand"
	"time"
	_ "time/tzdata"

	"github.com/bwmarrin/discordgo"

	"github.com/pipexlul/conceal-bot-go/internal/commands"
	"github.com/pipexlul/conceal-bot-go/internal/mappers"
	"github.com/pipexlul/conceal-bot-go/internal/models"
	concealbot "github.com/pipexlul/conceal-bot-go/internal/pkg/conceal-bot"
	"github.com/pipexlul/conceal-bot-go/internal/pkg/discord-status"
	"github.com/pipexlul/conceal-bot-go/internal/pkg/env"
	"github.com/pipexlul/conceal-bot-go/internal/pkg/server"
	"github.com/pipexlul/conceal-bot-go/internal/types"
	utils "github.com/pipexlul/conceal-bot-go/internal/utilities"
)

func main() {
	var err error

	concealBot := concealbot.NewConcealBot(
		concealbot.WithRandGen(rand.New(rand.NewSource(time.Now().UnixNano()))),
	)

	concealBot.ConnectMongo()
	defer concealBot.DisconnectMongo()

	botToken := env.GetBotToken()
	if botToken.Token == "" {
		log.Fatal("Missing all BOT TOKEN environment variables, at least one is required")
	}

	log.Printf("Starting bot in %s mode", mappers.MapTokenTypeToString(botToken.TokenType))

	concealBot.DiscordClient, err = discordgo.New("Bot " + botToken.Token)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	concealBot.StatusHelper = discordstatus.New(concealBot.DiscordClient, concealBot.RandGen)

	concealBot.RegisterHandler(onReady, models.EventType_Ready)

	if err = concealBot.DiscordClient.Open(); err != nil {
		log.Fatalf("Failed to open Discord session: %v", err)
	}
	log.Println("Discord session opened :)")

	defer func() {
		closeErr := concealBot.DiscordClient.Close()
		if closeErr != nil {
			log.Fatalf("Failed to close Discord session: %v", closeErr)
		}
		log.Println("Discord session closed")
	}()

	registererCtx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()

	// TODO: Use a more centralized command registerer
	var (
		cmd            types.BotCommand
		registerCmdErr error
	)
	cmd = &commands.TimeDifferenceCmd{}
	registerCmdErr = cmd.Register(registererCtx, concealBot)
	if registerCmdErr != nil {
		log.Fatalf("Failed to register command: %v", cmd.GetCommandName())
	}

	cmd = &commands.SpoillessVideoCmd{}
	registerCmdErr = cmd.Register(registererCtx, concealBot)
	if registerCmdErr != nil {
		log.Fatalf("Failed to register command: %v", cmd.GetCommandName())
	}

	concealBot.Validate()

	log.Print("All commands registered!")

	log.Println("Bot is now running. Press CTRL-C to exit.")

	concealBot.StatusHelper.SetupStatusTicker()

	apiServer := &server.APIServer{}
	apiServer.Start(concealBot)
	select {}
}

func onReady(bot utils.ConcealBot, s *discordgo.Session, _ interface{}) {
	if err := bot.GetStatusHelper().UpdateStatusFromRandom(); err != nil {
		log.Printf("Error updating game status at ready: %v", err)
	}
	log.Printf("Ready! Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
}
