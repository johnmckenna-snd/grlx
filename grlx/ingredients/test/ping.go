package test

import (
	"bytes"
	"encoding/json"
	"net/http"

	. "github.com/gogrlx/grlx/config"
	. "github.com/gogrlx/grlx/types"
)

func FPing(id string) (bool, error) {
	var s Inline200
	url := FarmerURL + "/test/ping"
	km := KeyManager{SproutID: id}
	jw, _ := json.Marshal(km)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jw))
	if err != nil {
		return false, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	err = json.NewDecoder(resp.Body).Decode(&s)
	return s.Success, err
}
