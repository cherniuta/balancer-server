package rest

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type PingResponse struct {
	Replie string `json:"replie"`
}

func PingPongHandler(log *slog.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var response PingResponse
		response.Replie = "pong"

		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(response); err != nil {
			log.Error("cannot encode reply", "error", err)
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

}

type GreetingResponse struct {
	Greeting string `json:"greeting"`
}

func SayHelloHandler(log *slog.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		name := request.URL.Query().Get("name")

		if name == "" {
			writer.WriteHeader(http.StatusBadRequest)
			log.Debug("empty name\n")
			return
		}

		greeting := GreetingResponse{Greeting: fmt.Sprintf("Hello, %s!", name)}

		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(greeting); err != nil {
			log.Error("cannot encode reply", "error", err)
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
