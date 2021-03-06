// Package shandler For the Authentication Server
//
// Documentation for REST API
//
//	Schemes: http
//	BasePath: /
//	Version: 1.0.5
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
// swagger:meta
package shandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// @termsOfService http://swagger.io/terms/

// swagger:route DELETE / Anything EMPTY
// Default Handler for everything that is not a match
//
// responses:
//	404: ErrorMessage

// DefaultHandler is for handling everything
func DefaultHandler(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusNotFound)
	Body := r.URL.Path + " is not supported. Thanks for visiting!\n"
	fmt.Fprintf(rw, "%s", Body)
}

// swagger:route GET /* NULL
// Default Handler for endpoints used with incorrect HTTP request method
//
// responses:
//	404: ErrorMessage

// MethodNotAllowedHandler is executed when the HTTP method is not supported
func MethodNotAllowedHandler(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusNotFound)
	Body := "Method not allowed!\n"
	fmt.Fprintf(rw, "%s", Body)
}

// swagger:route GET /v1/time time NULL
// Return current time
//
// responses:
//	200: OK
//	400: BadRequest

// TimeHandler is for handling /v1/time
func TimeHandler(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
	t := time.Now().Format(time.RFC1123)
	Body := "The current time is: " + t + "\n"
	fmt.Fprintf(rw, "%s", Body)
}

// swagger:route POST /v1/add createUser Input
// Create a new user
//
// responses:
//	200: OK
//  400: BadRequest

// AddHandler is for adding a new user /v1/add
func AddHandler(rw http.ResponseWriter, r *http.Request) {
	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	if len(d) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var users = []Input{}
	err = json.Unmarshal(d, &users)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println(users)

	u := UserPass{users[0].Username, users[0].Password}
	if !IsUserAdmin(u) {
		log.Println("Command issued by non-admin user:", u.Username)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	newUser := User{-1, users[1].Username, users[1].Password, time.Now().Unix(), users[1].Admin, 0}
	result := AddUser(newUser)
	if !result {
		rw.WriteHeader(http.StatusBadRequest)
	}
}

// DeleteHandler is for deleting an existing user + DELETE
func DeleteHandler(rw http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		log.Println("ID value not set!")
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	var user = UserPass{}
	err := user.FromJSON(r.Body)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if !IsUserAdmin(user) {
		log.Println("User", user.Username, "is not admin!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("id", err)
		return
	}

	t := FindUserID(intID)
	if t.Username != "" {
		log.Println("About to delete:", t)
		deleted := DeleteUser(intID)
		if deleted {
			log.Println("User deleted:", id)
			rw.WriteHeader(http.StatusOK)
			return
		} else {
			log.Println("Cannot delete user:", id)
			rw.WriteHeader(http.StatusNotFound)
		}
	}
	rw.WriteHeader(http.StatusNotFound)
}

// GetAllHandler is for getting all data from the user database
func GetAllHandler(rw http.ResponseWriter, r *http.Request) {
	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	if len(d) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var user = UserPass{}
	err = json.Unmarshal(d, &user)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if !IsUserValid(user) {
		log.Println("User", user, "does not exist!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	err = SliceToJSON(ReturnAllUsers(), rw)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
}

// swagger:route GET /v1/getid Username-Password UserPass
// Create a new user
//
// responses:
//	200: OK
//  400: BadRequest

// GetIDHandler returns the ID of an existing user
func GetIDHandler(rw http.ResponseWriter, r *http.Request) {
	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	if len(d) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var user = UserPass{}
	err = json.Unmarshal(d, &user)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("Input user:", user)
	if !IsUserValid(user) {
		log.Println("User", user.Username, "not valid!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	t := FindUserUsername(user.Username)
	Body := "User " + user.Username + " has ID:"
	fmt.Fprintf(rw, "%s %d\n", Body, t.ID)
}

// GetUserDataHandler + GET returns the full record of a user
func GetUserDataHandler(rw http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		log.Println("ID value not set!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("id", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	t := FindUserID(intID)
	if t.Username != "" {
		err := t.ToJSON(rw)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}
	} else {
		log.Println("User not found:", id)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
}

// UpdateHandler is for updating the data of an existing user + PUT
func UpdateHandler(rw http.ResponseWriter, r *http.Request) {
	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	if len(d) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var users = []Input{}
	err = json.Unmarshal(d, &users)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	u := UserPass{users[0].Username, users[0].Password}
	if !IsUserAdmin(u) {
		log.Println("Command issued by non-admin user:", u.Username)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println(users)
	t := FindUserUsername(users[1].Username)
	t.Username = users[1].Username
	t.Password = users[1].Password
	t.Admin = users[1].Admin

	if !UpdateUser(t) {
		log.Println("Update failed:", t)
		rw.WriteHeader(http.StatusBadRequest)
	}
}

// swagger:route POST /v1/login Username-Password UserPass
// Create a new user
//
// responses:
//	200: OK
//  400: BadRequest

// LoginHandler is for updating the LastLogin time of a user
// And changing the Active field to true
func LoginHandler(rw http.ResponseWriter, r *http.Request) {
	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	if len(d) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var user = UserPass{}
	err = json.Unmarshal(d, &user)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("Input user:", user)

	if !IsUserValid(user) {
		log.Println("User", user.Username, "not valid!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	t := FindUserUsername(user.Username)
	log.Println("Logging in:", t)

	t.LastLogin = time.Now().Unix()
	t.Active = 1
	if UpdateUser(t) {
		log.Println("User updated:", t)
	} else {
		log.Println("Update failed:", t)
		rw.WriteHeader(http.StatusBadRequest)
	}
}

// swagger:route POST /v1/logout UserPass
// Create a new user
//
// responses:
//	200: OK
//  400: BadRequest

// LogoutHandler is for logging out a user
// And changing the Active field to false
func LogoutHandler(rw http.ResponseWriter, r *http.Request) {
	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	if len(d) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Println("No input!")
		return
	}

	var user = UserPass{}
	err = json.Unmarshal(d, &user)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if !IsUserValid(user) {
		log.Println("User", user.Username, "exists!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	t := FindUserUsername(user.Username)
	log.Println("Logging out:", t.Username)
	t.Active = 0
	if UpdateUser(t) {
		log.Println("User updated:", t)
	} else {
		log.Println("Update failed:", t)
		rw.WriteHeader(http.StatusBadRequest)
	}
}

// swagger:route GET /v1/logged UserPass
// Get a list of logged in users
//
// responses:
//	200: OK
//  400: BadRequest

// LoggedUsersHandler returns the list of currently logged in users
func LoggedUsersHandler(rw http.ResponseWriter, r *http.Request) {
	var user = UserPass{}
	err := user.FromJSON(r.Body)

	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if !IsUserValid(user) {
		log.Println("User", user.Username, "exists!")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	err = SliceToJSON(ReturnLoggedUsers(), rw)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
}
