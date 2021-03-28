package types

type Message struct {
	Subsystem string `json:"subsystem"`
	Message   string `json:"message"`
}
