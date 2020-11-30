package service

import (
	"net/http"
)

type HTTP struct {
	protocol string
	mux      *http.ServeMux
	handler  *Handler
}

func NewHTTP(PR *ProducerReceiver, CR *ConsumerReceiver) *HTTP {
	handler := &Handler{
		ProducerReceiver: PR,
		ConsumerReceiver: CR,
	}
	return &HTTP{
		protocol: "http",
		mux:      NewServeMux(handler),
	}
}

func (h *HTTP) Start() {
	_ = http.ListenAndServe(":8000", h.mux)
}
