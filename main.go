package main

import (
	"log"
	"log/slog"
	"net/http"
)

func main() {

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	slog.Info("server starting.....")

	log.Fatal(http.ListenAndServe(":5000", nil))
}
