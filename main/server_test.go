package main

import (
	"encoding/json"
	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/utils"
	"testing"
)

// TODO add more tests for the other features of "DoRequest"

func TestPing(t *testing.T) {
	client := &utils.ComponentClient{RootUrl: "http://localhost:8080"}
	respBody, err := client.DoRequest("/ping", nil, "")
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}

	var responseString utils.SingleString
	err = json.Unmarshal(respBody, &responseString)
	assert.Nil(t, err)
	assert.Equal(t, "pong", responseString.Value)
}
