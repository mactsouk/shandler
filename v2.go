package shandler

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// User defines the structure for the payload of V2 of the REST API
// swagger:model
type V2Input struct {
	// The Username of the user issuing the command
	//
	// required: true
	Username string `json:"username"`
	// The Password of the user issuing the command
	//
	// required: true
	Password string `json:"password"`
	// The User that the command will affect
	//
	// required: false
	U User `json:"load"`
}

// IMAGESPATH defines the path where binary files are stored
var IMAGESPATH string

// swagger:route POST /v2/add V2Input
// Create a new user
//
// responses:
//	200: OK
//  400: BadRequest

// AddHandlerV2 is for adding new users /v2/add
func AddHandlerV2(rw http.ResponseWriter, r *http.Request) {
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

	var load = V2Input{}
	err = json.Unmarshal(d, &load)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println(load)

	u := UserPass{load.Username, load.Password}
	if !IsUserAdmin(u) {
		log.Println("Command issued by non-admin user:", u.Username)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	newUser := load.U
	result := AddUser(newUser)
	if !result {
		rw.WriteHeader(http.StatusBadRequest)
	}
}

// swagger:route POST /v2/login V2Input
// Create a new user
//
// responses:
//	200: OK
//  400: BadRequest

func LoginHandlerV2(rw http.ResponseWriter, r *http.Request) {
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

	var load = V2Input{}
	err = json.Unmarshal(d, &load)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	var user = UserPass{load.Username, load.Password}
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

// swagger:route POST /v2/logout V2Input
// Create a new user
//
// responses:
//	200: OK
//  400: BadRequest

func LogoutHandlerV2(rw http.ResponseWriter, r *http.Request) {
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

	var load = V2Input{}
	err = json.Unmarshal(d, &load)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	var user = UserPass{load.Username, load.Password}
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

// swagger:route GET /v2/getall V2Input Users
// Get a list of all users
//
// responses:
//	200: User
//  400: BadRequest

func GetAllHandlerV2(rw http.ResponseWriter, r *http.Request) {
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

	var load = V2Input{}
	err = json.Unmarshal(d, &load)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	var user = UserPass{load.Username, load.Password}
	if !IsUserAdmin(user) {
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

// swagger:route GET /v1/getall UserPass Users
// Get a list of all users
//
// responses:
//	200: User
//  400: BadRequest

// GetAllHandlerUpdated is for `/v1/getall`.
// The older version had a bug as it was using `IsUserValid` instead of `IsUserAdmin`.
func GetAllHandlerUpdated(rw http.ResponseWriter, r *http.Request) {
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

	if !IsUserAdmin(user) {
		log.Println("User", user, "is not Admin!")
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	err = SliceToJSON(ReturnAllUsers(), rw)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
}

// swagger:route PUT /v2/files/{filename} NULL
// Upload a new file
//
// responses:
//	200: OK
//	404: BadRequest

// UploadFile is for uploading files to the server
func UploadFile(rw http.ResponseWriter, r *http.Request) {
	filename, ok := mux.Vars(r)["filename"]
	if !ok {
		log.Println("filename value not set!")
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	log.Println(filename)
	saveFile(IMAGESPATH+"/"+filename, rw, r)
}

func saveFile(path string, rw http.ResponseWriter, r *http.Request) {
	log.Println("Saving to", path)
	err := saveToFile(path, r.Body)
	if err != nil {
		log.Println(err)
		return
	}
}

func saveToFile(path string, contents io.Reader) error {
	_, err := os.Stat(path)
	if err == nil {
		err = os.Remove(path)
		if err != nil {
			log.Println("Error deleting", path)
			return err
		}
	} else if !os.IsNotExist(err) {
		log.Println("Unexpected error:", err)
		return err
	}

	// If everything is OK, create the file
	f, err := os.Create(path)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()

	n, err := io.Copy(f, contents)
	if err != nil {
		return err
	}
	log.Println("Bytes written:", n)
	return nil
}

func CreateImageDirectory(d string) error {
	_, err := os.Stat(d)
	if os.IsNotExist(err) {
		log.Println("Creating:", d)
		err = os.MkdirAll(d, 0755)
		if err != nil {
			log.Println(err)
			return err
		}
	} else if err != nil {
		log.Println(err)
		return err
	}

	fileInfo, err := os.Stat(d)
	mode := fileInfo.Mode()
	if !mode.IsDir() {
		msg := d + " is not a directory!"
		return errors.New(msg)
	}
	return nil
}

func MiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving %s from %s using %s method", r.RequestURI, r.Host, r.Method)
		next.ServeHTTP(w, r)
	})
}

// Generating Random Strings with a Given length
func random(min, max int) int {
	return rand.Intn(max-min) + min
}

// RandomPassword generates random strings of given length
func RandomPassword(l int) string {
	Password := ""
	rand.Seed(time.Now().Unix())
	MIN := 0
	MAX := 94
	startChar := "!"
	i := 1
	for {
		myRand := random(MIN, MAX)
		newChar := string(startChar[0] + byte(myRand))
		Password += newChar
		if i == l {
			break
		}
		i++
	}
	return Password
}
