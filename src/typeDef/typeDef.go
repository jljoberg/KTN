package typeDef

type ClientMsg struct {
	Request string
	Content string
}

type ServerMsg struct {
	Timestamp string
	Sender    string
	Response  string
	Content   string
}

type HistoryMsg struct {
	Timestamp string
	Sender    string
	Response  string
	Content   [][]byte
}
