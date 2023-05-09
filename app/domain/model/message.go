package model

type Message struct {
	Payload map[string]interface{}
}

func NewMessage(payload map[string]interface{}) *Message {
	return &Message{Payload: payload}
}
