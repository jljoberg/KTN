package typeDef

// ClientPayload exported
type ClientPayload struct {
	Request string `json:"request"`
	Content string `json:"content"`
}

// ServerPayload exported
type ServerPayload struct {
	Timestamp string `json:"timestamp"`
	Sender    string `json:"sender"`
	Response  string `json:"response"`
	Content   string `json:"content"`
}

// HistoryPayload Exported
type HistoryPayload struct {
	Timestamp string   `json:"timestamp"`
	Sender    string   `json:"sender"`
	Response  string   `json:"response"`
	Content   [][]byte `json:"content"`
}
