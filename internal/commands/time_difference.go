package commands

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type TimeDifferenceCmd struct {
}

func (cmd *TimeDifferenceCmd) GetCommandName() string {
	return "timediff"
}

func (cmd *TimeDifferenceCmd) GetOptions() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "time",
			Description: "Base time from where you want to convert to other timezones",
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildText,
			},
			Required: true,
		},
	}
}

func (cmd *TimeDifferenceCmd) Register(
	_ context.Context,
	dgSession *discordgo.Session,
) error {
	// I estimate registering a command will not take over 6 seconds
	// No need for context usage at the moment, so I'll leave this commented
	// registerCtx, cancel := context.WithDeadline(ctx, time.Now().Add(6*time.Second))
	// defer cancel()

	// TODO: Maybe add the result to a map of sorts to unregister when needed?
	_, err := dgSession.ApplicationCommandCreate(dgSession.State.User.ID, "", &discordgo.ApplicationCommand{
		ID:          cmd.GetCommandName(),
		Type:        discordgo.ChatApplicationCommand,
		Name:        cmd.GetCommandName(),
		Description: "Convert any Time (HH:MM [AM/PM/am/pm]) to all other conceal timezones",
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

func (cmd *TimeDifferenceCmd) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	const (
		timeFormat24     = "15:04"
		timeFormat12     = "03:04PM"
		timeFormatOutput = "03:04 PM"
	)

	timeStr := i.ApplicationCommandData().Options[0].StringValue()
	timeStr = strings.ToUpper(timeStr)

	// Regex validations to add leading zeroes when necessary
	re := regexp.MustCompile(`^(\d):`)
	timeStr = re.ReplaceAllString(timeStr, "0$1:")
	re = regexp.MustCompile(`:(\d)(?:\s|AM|PM|$)`)
	timeStr = re.ReplaceAllString(timeStr, ":0$1")

	timeStr = strings.ReplaceAll(timeStr, " ", "")

	var (
		parsedTime    time.Time
		parsedTimeErr error
	)

	if strings.Contains(timeStr, "PM") || strings.Contains(timeStr, "AM") {
		parsedTime, parsedTimeErr = time.Parse(timeFormat12, timeStr)
	} else {
		parsedTime, parsedTimeErr = time.Parse(timeFormat24, timeStr)
	}

	if parsedTimeErr != nil {
		replyWithMessage(
			s,
			i,
			"Invalid time format. Please use one of the following formats: HH:MM (24-hour), HH:MM AM (HH:MM am), or HH:MM PM (HH:MM pm).",
		)
		return
	}

	timezonesMap := map[string]string{
		"New Jersey/Philadelphia": "America/New_York",
		"Chile":                   "America/Santiago",
	}

	var results []string
	for locationName, timezone := range timezonesMap {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			log.Printf("Error loading location %s: %v", timezone, err)
			continue
		}

		now := time.Now()
		timeInLocation := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			parsedTime.Hour(),
			parsedTime.Minute(),
			0,
			0,
			loc,
		)
		results = append(results,
			fmt.Sprintf(
				"If %s is the time in %s, then:",
				parsedTime.Format(timeFormatOutput),
				locationName,
			),
		)

		for targetName, targetTimezone := range timezonesMap {
			if targetName == locationName {
				continue
			}
			targetLoc, err := time.LoadLocation(targetTimezone)
			if err != nil {
				log.Printf("Error loading location %s: %v", targetTimezone, err)
				continue
			}
			targetTime := timeInLocation.In(targetLoc)
			results = append(results,
				fmt.Sprintf(
					"- %s would be the time in %s",
					targetTime.Format(timeFormatOutput),
					targetName,
				),
			)
		}
		results = append(results, "")
	}
	replyWithMessage(s, i, strings.Join(results, "\n"))
}