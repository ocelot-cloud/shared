package secutils

import (
	"encoding/json"
	"github.com/ocelot-cloud/shared"
	"net/http"
	"os"
)

var Logger = shared.ProvideLogger(os.Getenv("LOG_LEVEL"))

const OriginHeader = "Origin"

type SingleString struct {
	Value string `json:"value"`
}

type ChangePasswordForm struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func SendJsonResponse(w http.ResponseWriter, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		Logger.Error("unmarshalling failed: %v", err)
		http.Error(w, "failed to prepare response data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
