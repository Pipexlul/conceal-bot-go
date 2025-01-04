package commands

import (
	"context"
	"errors"
	"github.com/pipexlul/conceal-bot-go/internal/pkg/env"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	spoillessvideo "github.com/pipexlul/conceal-bot-go/internal/commands/specialized/spoilless_video"
	"github.com/pipexlul/conceal-bot-go/internal/types"
	utils "github.com/pipexlul/conceal-bot-go/internal/utilities"
)

type SpoillessVideoCmd struct {
	ConcealBot       utils.ConcealBot
	embedsCollection *mongo.Collection
}

var _ types.BotCommand = (*SpoillessVideoCmd)(nil)

func (cmd *SpoillessVideoCmd) InitMongoCollection() {
	cmd.embedsCollection = cmd.ConcealBot.GetMongoClient().Database("command_data").Collection("embeds")
}

func (cmd *SpoillessVideoCmd) Register(_ context.Context, bot utils.ConcealBot) error {
	cmd.ConcealBot = bot
	cmd.InitMongoCollection()

	dgSession := cmd.ConcealBot.Client()

	_, err := dgSession.ApplicationCommandCreate(dgSession.State.User.ID, "", &discordgo.ApplicationCommand{
		ID:          cmd.GetCommandName(),
		Type:        discordgo.ChatApplicationCommand,
		Name:        cmd.GetCommandName(),
		Description: "Embeds a youtube video by changing the title and optionally hiding the thumbnail",
		Options:     cmd.GetOptions(),
	})
	if err != nil {
		return err
	}

	dgSession.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand && i.ApplicationCommandData().Name == cmd.GetCommandName() {
			cmd.Execute(s, i)
		}
	})

	return nil
}

func (cmd *SpoillessVideoCmd) GetCommandName() string {
	return "spoilless-vid"
}

func (cmd *SpoillessVideoCmd) GetOptions() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "url",
			Description: "Youtube Video URL",
			Type:        discordgo.ApplicationCommandOptionString,
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildText,
			},
			Required: true,
		},
		{
			Name:        "new_title",
			Description: "Title to show instead of actual video title",
			Type:        discordgo.ApplicationCommandOptionString,
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildText,
			},
			Required: true,
		},
		{
			Name:        "hide_thumbnail",
			Description: "Hide video thumbnail (will show as black)",
			Type:        discordgo.ApplicationCommandOptionBoolean,
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildText,
			},
			Required: false,
		},
	}
}

func (cmd *SpoillessVideoCmd) Execute(dgSession *discordgo.Session, i *discordgo.InteractionCreate) {
	if !cmd.validateYoutubeLink(i.ApplicationCommandData().Options[0].StringValue()) {
		if err := dgSession.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid YouTube URL",
			},
		}); err != nil {
			log.Printf("[ERROR] Failed to respond to interaction '%s': %v", cmd.GetCommandName(), err)
		}
		return
	}

	if err := dgSession.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		log.Printf("[ERROR] Failed to defer respond to interaction '%s': %v", cmd.GetCommandName(), err)
		return
	}

	go func() {
		finalURL, err := cmd.handleEmbedCmd(i)
		if err != nil {
			log.Printf("[ERROR] Failed to handle interaction '%s': %v", cmd.GetCommandName(), err)
			return
		}

		log.Printf("[INFO] Successfully handled interaction '%s'", cmd.GetCommandName())
		if _, err := dgSession.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &finalURL,
		}); err != nil {
			log.Printf("[ERROR] Failed to respond to interaction '%s': %v", cmd.GetCommandName(), err)
		}
	}()
}

func (cmd *SpoillessVideoCmd) handleEmbedCmd(interaction *discordgo.InteractionCreate) (string, error) {
	var embed *spoillessvideo.Embed

	opts := interaction.ApplicationCommandData().Options

	videoURL := opts[0].StringValue()
	customTitle := opts[1].StringValue()
	hideThumbnail := len(opts) > 2 && opts[2].BoolValue()

	videoID, err := ExtractYoutubeVideoID(videoURL)
	if err != nil {
		return "", err
	}

	findQuery := bson.M{
		"video_id":     videoID,
		"custom_title": customTitle,
	}

	// TODO: Improve context usage
	err = cmd.embedsCollection.FindOne(context.Background(), findQuery).Decode(&embed)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return "", err
	}

	queryParams := url.Values{}
	queryParams.Add("video", videoID)
	queryParams.Add("title", customTitle)
	if hideThumbnail {
		queryParams.Add("hide_thumbnail", "true")
	}

	if embed == nil {
		embed = &spoillessvideo.Embed{
			VideoID:     videoID,
			CustomTitle: customTitle,
			Metadata: spoillessvideo.Metadata{
				Description: "Spoilless Video! - Click to watch on youtube",
				OGTitle:     customTitle,
				OGURL:       videoURL,
			},
			CreatedAt: time.Now(),
		}

		if _, err := cmd.embedsCollection.InsertOne(context.Background(), embed); err != nil {
			return "", err
		}
	}

	embedUrl := url.URL{
		Scheme:   "http",
		Host:     env.GetString("HOSTNAME", "localhost"),
		Path:     "/embed",
		RawQuery: queryParams.Encode(),
	}

	return embedUrl.String(), nil
}

func (cmd *SpoillessVideoCmd) validateYoutubeLink(link string) bool {
	validUrls := []string{
		"youtube.com",
		"youtu.be",
	}

	for _, validUrl := range validUrls {
		if strings.Contains(link, validUrl) {
			return true
		}
	}
	return false
}
