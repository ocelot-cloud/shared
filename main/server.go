package main

import (
	"github.com/ocelot-cloud/shared/utils"
	"net/http"
	"os"
	"os/exec"
)

type SingleString struct {
	Value string `json:"value"`
}

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		data := SingleString{Value: "pong"}
		utils.SendJsonResponse(w, data)
	})
	utils.Logger.InfoF("Listening on port 8080")
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			utils.Logger.ErrorF("ErrorF starting server: %v", err)
			os.Exit(1)
		}
	}()

	err := exec.Command("go", "test", "-v", "./...").Run()
	if err != nil {
		utils.Logger.ErrorF("Tests failed: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
