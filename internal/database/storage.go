package database

type IStorage interface {
	Disconnect()
	IsConnected() bool

	AddMessage(msg *Message) (err error)
	AddRecipient(id int) (err error)

	RemoveRecipient(id int) (err error)

	GetRecipients() (rcpns []Recipient, err error)
}
