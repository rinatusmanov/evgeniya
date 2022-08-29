package types

type GetBalanceRequest struct {
	CallerID             int    `json:"callerId"`
	PlayerName           string `json:"playerName"`
	Currency             string `json:"currency"`
	GameID               string `json:"gameId"`
	SessionID            string `json:"sessionId"`
	SessionAlternativeID string `json:"sessionAlternativeId"`
	BonusID              string `json:"bonusId"`
}

type GetBalanceResponse struct {
	Balance        int `json:"balance"`
	FreeRoundsLeft int `json:"freeroundsLeft"`
}
