package datamodel

type Message struct {
	ID      string
	Subject string
	From    string
	To      string
	Body    string
	Payload []byte
}
