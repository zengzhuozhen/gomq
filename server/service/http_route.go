package service

import "net/http"

func NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/version",Version)
	mux.HandleFunc("/messages", Messages)

	return mux
}