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