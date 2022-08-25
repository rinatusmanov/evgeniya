package types

type RollbackTransactionRequest struct {
	CallerID             int    `json:"callerId"`
	PlayerName           string `json:"playerName"`
	TransactionRef       string `json:"transactionRef"`
	GameID               string `json:"gameId"`
	SessionID            string `json:"sessionId"`
	SessionAlternativeID string `json:"sessionAlternativeId"`
}

type RollbackTransactionResponse struct{}
