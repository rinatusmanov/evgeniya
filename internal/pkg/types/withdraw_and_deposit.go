package types

type WithdrawAndDepositRequest struct {
	CallerID             int         `json:"callerId"`
	PlayerName           string      `json:"playerName"`
	Withdraw             int         `json:"withdraw"`
	Deposit              int         `json:"deposit"`
	Currency             string      `json:"currency"`
	TransactionRef       string      `json:"transactionRef"`
	GameRoundRef         string      `json:"gameRoundRef"`
	GameID               string      `json:"gameId"`
	Reason               Reason      `json:"reason"`
	SessionID            string      `json:"sessionId"`
	SessionAlternativeID string      `json:"sessionAlternativeId"`
	ChargeFreeRounds     int         `json:"chargeFreerounds"`
	BonusID              string      `json:"bonusId"`
	SpinDetails          SpinDetails `json:"spinDetails"`
}

type SpinDetails struct {
	BetType string `json:"betType"`
	WinType string `json:"winType"`
}

type Reason string

const (
	GamePlay      Reason = "GAME_PLAY"
	GamePlayFinal Reason = "GAME_PLAY_FINAL"
)

type WithdrawAndDepositResponse struct {
	NewBalance     int    `json:"newBalance"`
	TransactionID  string `json:"transactionId"`
	FreeRoundsLeft int    `json:"freeroundsLeft"`
}
