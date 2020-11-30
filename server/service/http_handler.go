package service

import (
	"encoding/json"
	"fmt"
	"gomq/cmd/do"
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
	vars := request.URL.Query();
	topicName := vars.Get("topic")
	fmt.Println(topicName)
	listMessageDo := new(do.MessagesDo)
	listMessageDo.TopicName = topicName

	for _, message := range h.ProducerReceiver.Queue.local[topicName] {
		listMessageDo.MessageList = append(listMessageDo.MessageList, message.Data.Body)
	}

	respData, _ := json.Marshal(listMessageDo)
	writer.Write(respData)

}

func (h *Handler) Publish(writer http.ResponseWriter, request *http.Request) {
	vars := request.URL.Query();
	topicName := vars.Get("topic")
	body := vars.Get("body")

	fmt.Println(topicName,body)

	writer.Write([]byte("ok"))
}
