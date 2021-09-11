package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/satori/go.uuid"
)

type Credentials struct {
    Password string `json:"password"`
    Username string `json:"username"`
}

var users = map[string]string{
    "user1": "password1",
    "user2": "password2",
}

func Singin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hello world")

    var creds Credentials

    err := json.NewDecoder(r.Body).Decode(&creds)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    expectedPassword, ok := users[creds.Username]

    if !ok || expectedPassword != creds.Password {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    sessionToken := uuid.NewV4().String()

    _, err = cache.Do("SETEX", sessionToken, "120", creds.Username)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    http.SetCookie(w, &http.Cookie {
        Name: "session_token",
        Value: sessionToken,
        Expires: time.Now().Add(120 * time.Second),
    })
}

func Welcome(w http.ResponseWriter, r *http.Request) {

    c, err := r.Cookie("session_token")
    if err != nil {
        if err == http.ErrNoCookie {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }

        w.WriteHeader(http.StatusBadRequest)
        return
    }

    sessionToken := c.Value

    response, err := cache.Do("GET", sessionToken)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    if response == nil {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    w.Write([]byte(fmt.Sprintf("Welcome %s!", response)))
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionToken := c.Value

	response, err := cache.Do("GET", sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if response != nil {
		w.WriteHeader(http.StatusUnauthorized)
	}

	newSessionToken := uuid.NewV4().String()
	_, err = cache.Do("SETEX", newSessionToken, "120", fmt.Sprintf("%s", response))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = cache.Do("DEL", sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie {
		Name: "session_token",
		Value: newSessionToken,
		Expires: time.Now().Add(120 * time.Second),
	})
}