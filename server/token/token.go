package token

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"
)

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func generateRandomString(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

func SetToken(w http.ResponseWriter, username string, sessions *map[string]string) {
	token, _ := generateRandomString(32)
	cookie := &http.Cookie{Name: "token", Value: token, Expires: time.Now().Add(30 * time.Minute)}
	http.SetCookie(w, cookie)
	(*sessions)[username] = token
}
