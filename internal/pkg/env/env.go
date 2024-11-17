package env

import (
	"log"
	"os"
	"strconv"

	"github.com/pipexlul/conceal-bot-go/internal/models"
)

func GetString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Error parsing %s: %v", key, err)
		return defaultValue
	}
	return intValue
}

func GetBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("Error parsing %s: %v", key, err)
		return defaultValue
	}
	return boolValue
}

func GetBotToken() models.BotToken {
	if token := os.Getenv("BOT_TOKEN"); token != "" {
		return models.BotToken{
			Token:     token,
			TokenType: models.BotTokenType_Production,
		}
	} else if token := os.Getenv("BOT_TOKEN_DEV"); token != "" {
		return models.BotToken{
			Token:     token,
			TokenType: models.BotTokenType_Development,
		}
	}

	// No valid envvars found, return an empty struct to the caller and let it handle the error
	return models.BotToken{}
}
