package service

import "net/http"

func NewServeMux(handler *Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/version",handler.Version)
	mux.HandleFunc("/messages", handler.Messages)
	mux.HandleFunc("/pub",handler.Publish)
	return mux
}

