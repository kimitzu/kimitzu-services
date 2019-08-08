package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/levigross/grequests"
)

type AuthPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type DjaliInfoP struct {
	Repo     string `json:"repoPath"`
	Cookie   string `json:"cookie"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func getInfo() (DjaliInfoP, error) {
	res, err := grequests.Get("http://127.0.0.1:4002/djali/info", &grequests.RequestOptions{RequestTimeout: time.Second * 10})
	if err != nil {
		fmt.Println("Error", err)
		return DjaliInfoP{}, fmt.Errorf("Can't resolve node, probably offline")
	}

	info := DjaliInfoP{}
	json.Unmarshal(res.Bytes(), &info)
	return info, nil

}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if r.Method != "POST" {
		http.Error(w, `{"error": "MethodNotPOST"}`, 405)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	payload := &AuthPayload{}
	err = json.Unmarshal(b, &payload)

	h := sha256.Sum256([]byte(payload.Password))
	password := hex.EncodeToString(h[:])

	info, err := getInfo()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if !(info.Username == payload.Username) || !(info.Password == password) {
		http.Error(w, `{"error": "Invalid credentials"}`, 403)
		return
	}

	fmt.Fprintf(w, `{"success": "%v"}`, info.Cookie)
}
