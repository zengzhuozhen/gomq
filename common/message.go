package common

type Message struct{
	MsgId int64 `json:"msg_id"`
	MsgKey string `json:"msg_key"`
	Body string `json:"body"`
}

func NewMessage(msgId int64, msgKey,body string) Message{
	return Message{
		MsgId: msgId ,
		MsgKey: msgKey,
		Body:   body,
	}
}