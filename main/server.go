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
	utils.Logger.Info("Listening on port 8080")
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			utils.Logger.Error("ErrorF starting server", utils.ErrorField, err)
			os.Exit(1)
		}
	}()

	err := exec.Command("go", "test", "-v", "./...").Run()
	if err != nil {
		utils.Logger.Error("Tests failed", utils.ErrorField, err)
		os.Exit(1)
	}

	os.Exit(0)
}
