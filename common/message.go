package common

type Message struct{
	MsgId int64
	MsgKey string
	Body string
}

func NewMessage(msgId int64, msgKey,body string) Message{
	return Message{
		MsgId: msgId ,
		MsgKey: msgKey,
		Body:   body,
	}
}