package service

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/zengzhuozhen/gomq/cmd/do"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/protocol/packet"
	"io/ioutil"
	"net/http"
)

type Handler struct {
	*ProducerReceiver
	*ConsumerReceiver
}

func (h *Handler) Version(writer http.ResponseWriter, request *http.Request) {
	versionDo := new(do.VersionDo)
	versionDo.GomqCtl = "v1"
	versionDo.Gomq = "v1"

	respData, _ := json.Marshal(versionDo)
	writer.Write(respData)
}

func (h *Handler) Messages(writer http.ResponseWriter, request *http.Request) {
	vars := request.URL.Query()
	topicName := vars.Get("topic")
	fmt.Println(topicName)
	listMessageDo := new(do.MessagesDo)
	listMessageDo.TopicName = topicName

	for _, message := range h.ProducerReceiver.RetainQueue.ReadAll(topicName) {
		listMessageDo.MessageList = append(listMessageDo.MessageList, message.Data.Body)
	}

	respData, _ := json.Marshal(listMessageDo)
	writer.Write(respData)

}

func (h *Handler) Publish(writer http.ResponseWriter, request *http.Request) {
	var data []byte
	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		writer.Write([]byte("err"))
		return
	}
	type ReqMessage struct {
		Topic  string `json:"topic"`
		Body   string `json:"body"`
		QoS    int32  `json:"qos"`
		Retain int32  `json:"retain"`
	}
	reqMessage := new(ReqMessage)
	_ = json.Unmarshal(data, reqMessage)
	fmt.Println(string(data))
	message := common.Message{
		MsgKey: uuid.New().String(),
		Body:   reqMessage.Body,
	}
	publishPacket := packet.NewPublishPacket(reqMessage.Topic, message, true, int(reqMessage.QoS), int(reqMessage.Retain), 0)
	h.ProducerReceiver.toQueue(&publishPacket)

	writer.Write([]byte("ok"))
}
