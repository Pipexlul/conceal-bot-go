package commands

import (
	"errors"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func ReplyWithMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
	if err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

func ExtractYoutubeVideoID(link string) (string, error) {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	switch parsedURL.Host {
	case "youtube.com", "www.youtube.com":
		id := parsedURL.Query().Get("v")
		if id == "" {
			return "", errors.New("video ID not found")
		}
		return id, nil
	case "youtu.be":
		id := strings.Trim(parsedURL.Path, "/")
		idRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`)

		if !idRegex.MatchString(id) {
			return "", errors.New("invalid video ID")
		}
		return id, nil
	default:
		return "", errors.New("invalid YouTube URL")
	}
}
