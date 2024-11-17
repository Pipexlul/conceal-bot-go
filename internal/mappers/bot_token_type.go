package mappers

import (
	"github.com/pipexlul/conceal-bot-go/internal/models"
)

var (
	tokenTypeToString = map[models.BotTokenType]string{
		models.BotTokenType_Unknown:     "Unknown",
		models.BotTokenType_Production:  "Production",
		models.BotTokenType_Development: "Development",
	}

	stringToTokenType = map[string]models.BotTokenType{
		"Unknown":     models.BotTokenType_Unknown,
		"Production":  models.BotTokenType_Production,
		"Development": models.BotTokenType_Development,
	}
)

func MapTokenTypeToString(tokenType models.BotTokenType) string {
	return tokenTypeToString[tokenType]
}

func MapStringToTokenType(tokenType string) models.BotTokenType {
	return stringToTokenType[tokenType]
}
