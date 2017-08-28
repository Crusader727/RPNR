package controller

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"../../models/user"
	"../token"
)

const solt = "i_love_perl"

func LoginUser(w http.ResponseWriter, r *http.Request, sessions *map[string]string) error {
	password := base64.URLEncoding.EncodeToString([]byte(r.FormValue("password") + solt))
	fmt.Println(password)
	fmt.Println("user" + r.FormValue("username"))
	u := user.User{Username: r.FormValue("username")}
	u.GetUserByName()
	fmt.Println(u.Password)
	if password == u.Password {
		expiration := time.Now().Add(24 * time.Hour)
		cookie := &http.Cookie{Name: "username", Value: u.Username, Expires: expiration}
		http.SetCookie(w, cookie)
		token.SetToken(w, u.Username, sessions)
		http.Redirect(w, r, "/", 302)
		return nil
	}
	return errors.New("login failed")
}

func RegisterUser(w http.ResponseWriter, r *http.Request) error {
	if r.FormValue("password") != r.FormValue("confirm") {
		return errors.New("Passwords don't match")
	}
	password := base64.URLEncoding.EncodeToString([]byte(r.FormValue("password") + solt))
	u := user.User{Username: r.FormValue("username"), Password: password}
	err := u.Create()
	if err != nil {
		return err
	}
	http.Redirect(w, r, "/login", 302)
	return nil
}

func GetValueFromCookie(val string, r *http.Request) (string, error) {
	cookie, err := r.Cookie(val)
	if err != nil {
		return "", err
	} else {
		return cookie.Value, nil
	}
}
