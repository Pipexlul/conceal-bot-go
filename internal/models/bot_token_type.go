package models

type BotTokenType int

const (
	BotTokenType_Unknown BotTokenType = iota
	BotTokenType_Production
	BotTokenType_Development
)
