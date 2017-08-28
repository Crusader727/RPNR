package server

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"../confreader"
	"../contexts/loggedin"
	"../models/history"
	"../rpnr"
	"../workers"
	"./controller"
)

type serverConfig struct {
	Server  string
	Workers int
}

//IPool - pool of routines
type IPool interface {
	Size() int
	Run()
	AddTaskSyncTimed(f workers.Func, timeout time.Duration) (interface{}, error)
}

const requestWaitInQueueTimeout = time.Millisecond * 100

//html templates
var hist = template.Must(template.ParseFiles("assets/history.html"))
var index = template.Must(template.ParseFiles("assets/index.html"))
var login = template.Must(template.ParseFiles("assets/login.html"))
var registration = template.Must(template.ParseFiles("assets/registration.html"))

//very important stuff
var sessions = make(map[string]string)
var conf serverConfig
var wp IPool

func loginHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		if r.Method == http.MethodGet {
			login.Execute(w, nil)
		} else {
			err := controller.LoginUser(w, r, &sessions)
			if err != nil {
				login.Execute(w, nil)
			} else {
				index.Execute(w, nil)
			}
		}
		return nil
	}, requestWaitInQueueTimeout)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
	}
}

func regHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		if r.Method == http.MethodGet {
			registration.Execute(w, nil)
		} else {
			err := controller.RegisterUser(w, r)
			if err != nil {
				log.Println(err)
				registration.Execute(w, struct{ Res string }{Res: ""})
			} else {
				http.Redirect(w, r, "/login", 302)
			}
		}
		return nil
	}, requestWaitInQueueTimeout)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
	}
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		u, ok := loggedin.FromContext(r.Context())
		if !ok {
			http.Redirect(w, r, "/login", 302)
			return nil
		}
		if r.Method == http.MethodGet {
			var notations []history.Notation
			history.GetLatest(u.ID, 10, &notations)
			hist.Execute(w, struct {
				Res      string
				Username string
				Srcs     []history.Notation
			}{Res: "", Srcs: notations, Username: u.Username})
		}
		return nil
	}, requestWaitInQueueTimeout)

	if err != nil {
		http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		u, ok := loggedin.FromContext(r.Context())
		if !ok {
			http.Redirect(w, r, "/login", 302)
			return nil
		}
		if r.Method == http.MethodGet {

			index.Execute(w, struct{ Res, Username string }{Res: "", Username: u.Username})

		} else {
			err := r.ParseMultipartForm(32 << 20)
			if err != nil {
				index.Execute(w, struct{ Res, Username, Err string }{Username: u.Username, Err: err.Error()})
			}
			file, handler, err := r.FormFile("uploadfile")
			if err != nil {
				fmt.Println(err)
				index.Execute(w, struct{ Res, Username, Err string }{Username: u.Username, Err: err.Error()})
				return nil
			}
			defer file.Close()

			handler.Filename = "assets/dbImages/" + u.Username + "_" + strconv.Itoa(time.Now().Nanosecond()) + ".png"
			f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				fmt.Println(err)
				index.Execute(w, struct{ Res, Username, Err string }{Username: u.Username, Err: err.Error()})
				return nil
			}
			defer f.Close()

			_, err = io.Copy(f, file)
			if err != nil {
				log.Println(err)
				http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
			}
			result, resolution, err := rpnr.GetPlateNumber(handler.Filename)
			if err != nil {
				index.Execute(w, struct{ Res, Username, Err string }{Res: "Result: " + err.Error()})
				return nil
			}
			index.Execute(w, struct{ Res, Username, Err string }{Res: "Result: " + result})
			n := history.Notation{
				Number: result,
				Img:    handler.Filename,
				ImgRes: resolution,
				UserID: u.ID,
			}
			err = n.Save()
			if err != nil {
				log.Println(err)
				http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
			}
		}
		return nil
	}, requestWaitInQueueTimeout)

	if err != nil {
		http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		u, ok := loggedin.FromContext(r.Context())
		if !ok {
			http.Redirect(w, r, "/login", 302)
			return nil
		}
		delete(sessions, u.Username)
		/*expiration := time.Now().AddDate(0, 0, -1)
		cookie := &http.Cookie{Name: "username", Value: u.Username, Expires: expiration}
		http.SetCookie(w, cookie)*/
		http.Redirect(w, r, "/login", 302)
		return nil
	}, requestWaitInQueueTimeout)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
	}
}

func init() {
	if err := confreader.ReadConfig("server", &conf); err != nil {
		fmt.Println(err)
	}
	fmt.Println(conf)
	wp = workers.NewPool(conf.Workers)
	wp.Run()
}

//RunHTTPServer - runs http server on address addr
func RunHTTPServer() error {
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
	mux.HandleFunc("/login", loginHandler)
	mux.Handle("/", loggedin.AddLoginContext(http.HandlerFunc(rootHandler), &sessions))
	mux.Handle("/logout", loggedin.AddLoginContext(http.HandlerFunc(logoutHandler), &sessions))
	mux.HandleFunc("/register", regHandler)
	mux.Handle("/history", loggedin.AddLoginContext(http.HandlerFunc(historyHandler), &sessions))
	return http.ListenAndServe(conf.Server, mux)
}
