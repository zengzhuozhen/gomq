package service

import (
	"encoding/json"
	"gomq/cmd/do"
	"net/http"
)


func Version(writer http.ResponseWriter, request *http.Request) {
	versionDo := new(do.VersionDo)
	versionDo.GomqCtl = "v1"
	versionDo.Gomq = "v1"

	respData ,_:=json.Marshal(versionDo)
	writer.Write(respData)
}

func Messages(writer http.ResponseWriter, request *http.Request){
	topicName := request.Form.Get("topic_name")

	listMessageDo :=  new(do.MessagesDo)
	listMessageDo.TopicName = topicName

	// todo

	respData ,_:=json.Marshal(listMessageDo)
	writer.Write(respData)

}


