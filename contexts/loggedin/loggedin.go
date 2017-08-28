package loggedin

import (
	"fmt"
	"net/http"

	"../../models/user"
	"../../server/controller"
	"../../server/token"

	"context"
)

type key string

const userLoginKey key = "loggedin"

// NewContext returns a new Context carrying userLogin.
func NewContext(ctx context.Context, u user.User) (context.Context, error) {
	return context.WithValue(ctx, userLoginKey, u), nil
}

//FromContext return context token
func FromContext(ctx context.Context) (user.User, bool) {
	user, ok := ctx.Value(userLoginKey).(user.User)
	return user, ok
}

func AddLoginContext(h http.Handler, sessions *map[string]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%+v", r)
		username, _ := controller.GetValueFromCookie("username", r)
		t, _ := controller.GetValueFromCookie("token", r)
		sessionToken, ok := (*sessions)[username]
		if sessionToken == t && ok {
			u := user.User{Username: username}
			u.GetUserByName()
			ctx, _ := NewContext(r.Context(), u)
			token.SetToken(w, u.Username, sessions)
			fmt.Println("with loggin context")
			h.ServeHTTP(w, r.WithContext(ctx))
		} else {
			fmt.Println("no loggin context")
			h.ServeHTTP(w, r)
		}
	})
}
