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
	Username    string `json:"username"`
	Password    string `json:"password"`
	NewUsername string `json:"newUsername"`
	NewPassword string `json:"newPassword"`
}

type KimitzuInfoP struct {
	Repo          string `json:"repoPath"`
	Cookie        string `json:"cookie"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Authenticated bool   `json:"authenticated"`
}

func getInfo() (KimitzuInfoP, error) {
    res, err := grequests.Get("http://127.0.0.1:4002/kimitzu/info", &grequests.RequestOptions{RequestTimeout: time.Second * 10})
	if err != nil {
		fmt.Println("Error", err)
        return KimitzuInfoP{}, fmt.Errorf("Can't resolve node, probably offline")
    }

    info := KimitzuInfoP{}
	json.Unmarshal(res.Bytes(), &info)
	return info, nil
}

func patchConfig(username, password string, authenticate bool) (KimitzuInfoP, error) {
    res, err := grequests.Post("http://127.0.0.1:4002/kimitzu/config", &grequests.RequestOptions{
		RequestTimeout: time.Second * 10,
		JSON: map[string]interface{}{
			"username":      username,
			"password":      password,
			"authenticated": authenticate,
		},
	})

	if err != nil {
		fmt.Println("Error", err)
        return KimitzuInfoP{}, fmt.Errorf("Can't resolve node, probably offline")
    }

    info := KimitzuInfoP{}
	json.Unmarshal(res.Bytes(), &info)
	return info, nil
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
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

	if r.Method == "GET" {
		fmt.Fprintf(w, `{"authentication": %v}`, info.Authenticated)
		return
	}

	if (!(info.Username == payload.Username) || !(info.Password == password)) && info.Authenticated {
		http.Error(w, `{"error": "Invalid credentials"}`, 403)
		return
	}

	if r.Method == "POST" {
		fmt.Fprintf(w, `{"method": "POST", "success": "%v"}`, info.Cookie)
		return
	}

	if r.Method == "PATCH" {
		if payload.NewUsername == "" || payload.NewPassword == "" {
			fmt.Fprintf(w, `{"method": "POST", "error": "new username or password is empty"}`)
			return
		}

		d, err := patchConfig(payload.NewUsername, payload.NewPassword, true)
		if err != nil {
			fmt.Fprintf(w, `{"method": "POST", "error": "%v"}`, err)
			return
		}

		fmt.Fprintf(w, `{"method": "PATCH", "success": "%v"}`, d)
		return
	}

	if r.Method == "DELETE" {
		d, err := patchConfig(payload.Username, payload.Password, false)
		if err != nil {
			fmt.Fprintf(w, `{"method": "DELETE", "error": "%v"}`, err)
			return
		}

		fmt.Fprintf(w, `{"method": "DELETE", "success": "%v"}`, d)
		return
	}

}
