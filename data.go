package shandler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/go-playground/validator"
	_ "github.com/mattn/go-sqlite3"
)

// SQLFILE defines the path of the SQLite3 database
var SQLFILE = "/tmp/users.db"

// User defines the structure for a Full User Record
// swagger:model User
type User struct {
	// The ID for the User
	// in: body
	//
	// required: false
	// min: 1
	ID int `json:"id"`
	// The Username of the User
	// in: body
	//
	// required: true
	Username string `json:"user"`
	// The Password of the User
	//
	// required: true
	Password string `json:"password"`
	// The Last Login time of the User
	//
	// required: true
	// min: 0
	LastLogin int64 `json:"lastlogin"`
	// Is the User Admin or not
	//
	// required: true
	Admin int `json:"admin"`
	// Is the User Logged In or Not
	//
	// required: true
	Active int `json:"active"`
}

// Input defines the structure for the user issuing a command
// swagger:model Input
type Input struct {
	// The Username of the User
	// in: body
	//
	// required: true
	Username string `json:"user"`
	// The Password of the User
	// in: body
	//
	// required: true
	Password string `json:"password"`
	// Is the User Admin or not
	//
	// required: true
	Admin int `json:"admin"`
}

// UserPass defines the structure for the user issuing a command
// swagger:model UserPass
type UserPass struct {
	// The Username of the User
	//
	// required: true
	Username string `json:"user" validate:"required"`
	// The Password of the User
	//
	// required: true
	Password string `json:"password" validate:"required"`
}

// Generic error message returned as an HTTP Status Code
// swagger:response ErrorMessage
type ErrorMessage struct {
	// Description of the situation
	// in: body
	Body int
}

// Generic OK message returned as an HTTP Status Code
// swagger:response OK
type OK struct {
	// Description of the situation
	// in: body
	Body int
}

// Generic BadRequest message returned an HTTP Status Code
// swagger:response BadRequest
type BadRequest struct {
	// Description of the situation
	// in: body
	Body int
}

const (
	empty = ""
	tab   = "\t"
)

// PrettyJSON is for pretty printing JSON records
func PrettyJSON(data interface{}) (string, error) {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent(empty, tab)

	err := encoder.Encode(data)
	if err != nil {
		return empty, err
	}
	return buffer.String(), nil
}

// FromJSON decodes a serialized JSON record - User{}
func (p *User) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(p)
}

// ToJSON serializes a JSON record
func (p *User) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

// FromJSON decodes a serialized JSON record - UserPass{}
func (p *UserPass) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(p)
}

// ToJSON serializes a JSON record
func (p *UserPass) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

// SliceFromJSON decodes a serialized slice with JSON records
func SliceFromJSON(slice interface{}, r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(slice)
}

// SliceToJSON serializes a slice with JSON records
func SliceToJSON(slice interface{}, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(slice)
}

// AddUser is for adding a new user to the database
func AddUser(u User) bool {
	log.Println("Adding user:", u)
	db, err := sql.Open("sqlite3", SQLFILE)
	if err != nil {
		log.Println(err)
		return false
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO users(Username, Password, LastLogin, Admin, Active) values(?,?,?,?,?)")
	if err != nil {
		log.Println("Adduser:", err)
		return false
	}
	stmt.Exec(u.Username, u.Password, u.LastLogin, u.Admin, u.Active)
	return true
}

// UpdateUser allows you to update user name
func UpdateUser(u User) bool {
	log.Println("Updating user:", u)

	db, err := sql.Open("sqlite3", SQLFILE)
	if err != nil {
		log.Println(err)
		return false
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE users SET Username=?, Password=?, LastLogin=?, Admin=?, Active =? WHERE ID = ?\n")
	if err != nil {
		log.Println("Statement failed:", err)
		return false
	}

	res, err := stmt.Exec(u.Username, u.Password, u.LastLogin, u.Admin, u.Active, u.ID)
	if err != nil {
		log.Println("Exec failed:", err)
		return false
	}

	affect, err := res.RowsAffected()
	log.Println("Affected:", affect)
	return true
}

// CreateDatabase initializes the SQLite3 database and adds the admin user
func CreateDatabase() bool {
	log.Println("Writing to SQLite3:", SQLFILE)
	db, err := sql.Open("sqlite3", SQLFILE)
	if err != nil {
		log.Println(err)
		return false
	}
	defer db.Close()

	log.Println("Emptying database table.")
	_, _ = db.Exec("DROP TABLE users")

	log.Println("Creating table from scratch.")
	_, err = db.Exec("CREATE TABLE users (ID integer NOT NULL PRIMARY KEY AUTOINCREMENT, Username TEXT, Password TEXT, Lastlogin integer, Admin integer, Active integer);")
	if err != nil {
		log.Println(err)
		return false
	}

	log.Println("Populating", SQLFILE)
	admin := User{-1, "admin", "admin", time.Now().Unix(), 1, 0}
	return AddUser(admin)
}

// DeleteUser is for deleting a user defined by ID
func DeleteUser(ID int) bool {
	log.Println("Deleting from SQLite3:", ID)
	db, err := sql.Open("sqlite3", SQLFILE)
	if err != nil {
		log.Println(err)
		return false
	}
	defer db.Close()

	stmt, _ := db.Prepare("DELETE FROM users WHERE ID = ?")
	if err != nil {
		log.Println("DeleteUser:", err)
		return false
	}
	stmt.Exec(ID)
	return true
}

// ReturnAllUsers is for returning all users from database
func ReturnAllUsers() []User {
	log.Println("Reading from SQLite3:", SQLFILE)
	db, err := sql.Open("sqlite3", SQLFILE)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users \n")
	if err != nil {
		log.Println(err)
		return nil
	}

	all := []User{}
	var c1 int
	var c2, c3 string
	var c4 int64
	var c5, c6 int

	for rows.Next() {
		err = rows.Scan(&c1, &c2, &c3, &c4, &c5, &c6)
		temp := User{c1, c2, c3, c4, c5, c6}
		all = append(all, temp)
	}

	log.Println("All:", all)
	return all
}

// FindUserID is for returning a user record defined by ID
func FindUserID(ID int) User {
	log.Println("Get User Data from SQLite3:", ID)
	db, err := sql.Open("sqlite3", SQLFILE)
	if err != nil {
		log.Println(err)
		return User{}
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users where ID = $1 \n", ID)
	if err != nil {
		log.Println("Query:", err)
		return User{}
	}
	defer rows.Close()

	u := User{}
	var c1 int
	var c2, c3 string
	var c4 int64
	var c5, c6 int

	for rows.Next() {
		err = rows.Scan(&c1, &c2, &c3, &c4, &c5, &c6)
		if err != nil {
			log.Println(err)
			return User{}
		}
		u = User{c1, c2, c3, c4, c5, c6}
		log.Println("Found user:", u)
	}
	return u
}

// FindUserUsername is for returning a user record defined by username
func FindUserUsername(username string) User {
	log.Println("Get User Data from SQLite3:", username)
	db, err := sql.Open("sqlite3", SQLFILE)
	if err != nil {
		log.Println(err)
		return User{}
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users where Username = $1 \n", username)
	if err != nil {
		log.Println("FindUserUsername Query:", err)
		return User{}
	}
	defer rows.Close()

	u := User{}
	var c1 int
	var c2, c3 string
	var c4 int64
	var c5, c6 int

	for rows.Next() {
		err = rows.Scan(&c1, &c2, &c3, &c4, &c5, &c6)
		if err != nil {
			log.Println(err)
			return User{}
		}
		u = User{c1, c2, c3, c4, c5, c6}
		log.Println("Found user:", u)
	}
	return u
}

// ReturnLoggedUsers is for returning all logged in users
func ReturnLoggedUsers() []User {
	log.Println("Reading from SQLite3:", SQLFILE)
	db, err := sql.Open("sqlite3", SQLFILE)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users WHERE Active = 1 \n")
	if err != nil {
		log.Println(err)
		return nil
	}

	all := []User{}
	var c1 int
	var c2, c3 string
	var c4 int64
	var c5, c6 int

	for rows.Next() {
		err = rows.Scan(&c1, &c2, &c3, &c4, &c5, &c6)
		if err != nil {
			log.Println(err)
			return []User{}
		}
		temp := User{c1, c2, c3, c4, c5, c6}
		log.Println("temp:", all)
		all = append(all, temp)
	}

	log.Println("Logged in:", all)
	return all
}

// IsUserAdmin determines whether a user is
// an administrator or not
func IsUserAdmin(u UserPass) bool {
	err := u.Validate()
	if err != nil {
		log.Println("IsUserAdmin - Validate:", err)
		return false
	}

	db, err := sql.Open("sqlite3", SQLFILE)
	if err != nil {
		log.Println(err)
		return false
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users WHERE Username = $1 \n", u.Username)
	if err != nil {
		log.Println(err)
		return false
	}

	temp := User{}
	var c1 int
	var c2, c3 string
	var c4 int64
	var c5, c6 int

	// If there exist multiple users with the same username,
	// we will get the FIRST ONE only.
	for rows.Next() {
		err = rows.Scan(&c1, &c2, &c3, &c4, &c5, &c6)
		if err != nil {
			log.Println(err)
			return false
		}
		temp = User{c1, c2, c3, c4, c5, c6}
	}

	if u.Username == temp.Username && u.Password == temp.Password && temp.Admin == 1 {
		return true
	}
	return false
}

func IsUserValid(u UserPass) bool {
	err := u.Validate()
	if err != nil {
		log.Println("IsUserValid - Validate:", err)
		return false
	}

	db, err := sql.Open("sqlite3", SQLFILE)
	if err != nil {
		log.Println(err)
		return false
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users WHERE Username = $1 \n", u.Username)
	if err != nil {
		log.Println(err)
		return false
	}

	temp := User{}
	var c1 int
	var c2, c3 string
	var c4 int64
	var c5, c6 int

	// If there exist multiple users with the same username,
	// we will get the FIRST ONE only.
	for rows.Next() {
		err = rows.Scan(&c1, &c2, &c3, &c4, &c5, &c6)
		if err != nil {
			log.Println(err)
			return false
		}
		temp = User{c1, c2, c3, c4, c5, c6}
	}

	if u.Username == temp.Username && u.Password == temp.Password {
		return true
	}
	return false
}

// Validate method validates the data of UserPass
func (p *UserPass) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}
