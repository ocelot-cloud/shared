package main

import (
	"github.com/ocelot-cloud/shared/utils"
	"net/http"
)

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		data := utils.SingleString{Value: "pong"}
		utils.SendJsonResponse(w, data)
	})
	utils.Logger.Info("Listening on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		utils.Logger.Error("Error starting server: %v", err)
		return
	}
}
