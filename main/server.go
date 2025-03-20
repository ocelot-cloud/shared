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
	println("Listening on port 8080")
	http.ListenAndServe(":8080", nil)
}
