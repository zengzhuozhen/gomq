package service

import (
	"net/http"
)

type HTTP struct {
	protocol string
	mux      *http.ServeMux
	*ProducerReceiver
	*ConsumerReceiver
}

func NewHTTP(PR *ProducerReceiver, CR *ConsumerReceiver) *HTTP {
	return &HTTP{
		protocol:         "http",
		mux:              NewServeMux(),
		ProducerReceiver: PR,
		ConsumerReceiver: CR,
	}
}



func (h *HTTP) Start() {
	_ = http.ListenAndServe(":8000", h.mux)
}
