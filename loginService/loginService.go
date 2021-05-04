package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// Resources: User Accounts
// Data to be passed: Usernames, hashed passwords, access keys
// Have its own database

// courseService will NOT interact with loginService, OTHER THAN TO CHECK ACCESS KEY through the validate() handler.
// Client will use authenticate function to verify the username+accesskey
// courseDB will have a column to keep track of the logged-in user, such that a logged-in user can only edit or delete courses created by themselves

// Non-Admin:
// Authenticate user (login): GET --> needs to return both bool and username+accesskey
// Create new user: POST (json)

// Admin:
// Authenticate admin (login) : GET -->  needs to return both bool and username+accesskey
// Create new user: POST (json)
// Edit existing user, incl. granting and revoking access keys (must be logged in as admin): PUT (json)
// Deleteing existing user: DELETE

type UserInfo struct {
	Username  string
	AccessKey string
}

type Account struct {
	Username string
	Password string
}

const (
	passwordHeader  = "TEAG*herd9tank-twis"
	accessKeyHeader = "keay*kak3jegh.BOB"
	usernameHeader  = "fer_ROUX9bam!preb"
)

var (
	db  *sql.DB
	err error // for sql.Open in init()
)

func init() {
	db, err = sql.Open("mysql", "root:veg-kluh!PRIW3hirt@tcp(127.0.0.1:3306)/login_db")
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Login database opened successfully.")
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/users/v1/", users)
	router.HandleFunc("/users/v1/{username}/{admin}", login).Methods("GET", "PUT", "POST", "DELETE")
	router.HandleFunc("/keys/v1/{accesskey}", validate).Methods("GET")

	fmt.Println("Listening at port 2000")
	log.Fatal(http.ListenAndServe(":2000", router))
}

// Handler Functions
func users(w http.ResponseWriter, r *http.Request) {
	// Print All users
	userrows, err := db.Query("SELECT username FROM login_db.login")
	if err != nil {
		panic(err)
	}
	var result []string
	for userrows.Next() {
		var user string
		err := userrows.Scan(&user)
		if err != nil {
			panic(err)
		}
		result = append(result, user)
	}
	fmt.Println(result)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func login(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fmt.Println("Params:", params)
	admin, err := strconv.ParseBool(params["admin"])
	if err != nil {
		panic(err.Error())
	}

	if r.Method == "GET" { // Authenticate user login
		fmt.Println("Login Authentication (non-admin) called")
		usernameInput, ok := params["username"]
		if !ok { // No username input
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Username and/or password is invalid or blank."))
			return
		}

		pwCookie, err := r.Cookie(passwordHeader)
		if err != nil { // No password input
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Username and/or password is invalid or blank."))
			return
		}
		pwInput := pwCookie.Value
		ok, user := getUser(db, usernameInput, pwInput, admin)

		if !ok {
			// User not found
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Username and/or password is invalid or blank"))
			return
		} else {
			// Login success, return response with username and accesskey
			usernameCookie := &http.Cookie{
				Name:  usernameHeader,
				Value: user.Username}
			accessKeyCookie := &http.Cookie{
				Name:  accessKeyHeader,
				Value: user.AccessKey}
			http.SetCookie(w, usernameCookie)
			http.SetCookie(w, accessKeyCookie)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("200 - User Authenticated:" + usernameInput))
			fmt.Println("login success")
		}
	}

	if r.Method == "DELETE" {
		usernameInput, ok := params["username"]
		if !ok { // No username input
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Username and/or password is invalid or blank."))
			return
		}

		// Check database if username exists
		result := db.QueryRow(fmt.Sprintf("SELECT * FROM login_db.login WHERE Username = '%s' LIMIT 1", usernameInput))
		if err := result.Scan(); err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - No user found"))
			return
		}

		deleteuser(db, usernameInput)
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("202 - User deleted: " + usernameInput))

	}

	if r.Method == "POST" { // Add new user without (course service access key to be further provisioned via admin client)
		var newAccount Account
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err.Error())
		}

		json.Unmarshal(reqBody, &newAccount)
		// Check database if duplicate
		result := db.QueryRow(fmt.Sprintf("SELECT * FROM login_db.login WHERE Username = '%s' LIMIT 1", newAccount.Username))
		if err := result.Scan(); err == nil {
			// Username already exists, return error for duplicate entry
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Duplicate entry detected."))
			return
		}

		// Create new username
		insertUser(db, newAccount, admin)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("201 - Account added : " + newAccount.Username))
	}
	if r.Method == "PUT" { // provision/revoke access key, only callable from Admin client
		// Search DB for whether access key exists
		result := db.QueryRow(fmt.Sprintf("SELECT * FROM login_db.login WHERE Username = '%s' LIMIT 1", params["username"]))
		if result.Err() != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - No user found"))
			return
		}

		resultScan := struct {
			Username  string
			Pw        string
			AccessKey string
		}{}

		result.Scan(&resultScan.Username, &resultScan.Pw, &resultScan.AccessKey)
		fmt.Println(resultScan)
		if resultScan.AccessKey == "nil" { // Check how driver handles null values
			//Generate UUID and assign as key. Print Status and outcome (assigned).
			accessKey := uuid.NewV4()

			edituserKey(db, params["username"], accessKey.String())
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("202 - Access Key provisioned: " + params["username"]))
		} else {
			// Remove Key. Print Status and outcome (revoked)
			edituserKey(db, params["username"], "nil")
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("202 - Access Key revoked: " + params["username"]))
		}
	}
}

func validate(w http.ResponseWriter, r *http.Request) {
	type Result struct {
		Validated bool
	}
	var result Result
	params := mux.Vars(r)
	accessKey, ok := params["accesskey"]
	if !ok { // No access key supplied
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("422 - No access key supplied."))
		return
	}
	DBsearch := db.QueryRow(fmt.Sprintf("SELECT * FROM login_db.login WHERE AccessKey = '%s' LIMIT 1", accessKey))
	if DBsearch.Err() != nil { // No row found
		w.WriteHeader(http.StatusNotFound)
		result.Validated = false
	} else {
		w.WriteHeader(http.StatusAccepted)
		result.Validated = true
	}
	json.NewEncoder(w).Encode(result)
}

// DB Operations
func getUser(db *sql.DB, username, pw string, admin bool) (bool, UserInfo) {
	account := db.QueryRow(fmt.Sprintf("Select * FROM login_db.login WHERE Username = '%s' LIMIT 1", username))
	if account.Err() != nil { // No match for username
		return false, UserInfo{"", ""}
	}
	var retrievedUser, retrievedPW, retrievedAccKey string
	err = account.Scan(&retrievedUser, &retrievedPW, &retrievedAccKey)
	if err != nil { // No match for username
		return false, UserInfo{"", ""}
	}
	if bcrypt.CompareHashAndPassword([]byte(retrievedPW), []byte(pw)) == nil {
		if !admin {
			return true, UserInfo{username, retrievedAccKey}
		} else {
			return true, UserInfo{username, ""}
		}
	} else {
		return false, UserInfo{"", ""}
	}
}

func insertUser(db *sql.DB, account Account, admin bool) {
	hashedPW, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.MinCost)

	var query string
	if err != nil {
		panic(err)
	}
	query = fmt.Sprintf("INSERT INTO login_db.login(Username, Pw, AccessKey) VALUES ('%s', '%s', '%s')", account.Username, string(hashedPW), "nil")
	_, err = db.Query(query)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func deleteuser(db *sql.DB, username string) {
	query := fmt.Sprintf("DELETE FROM login_db.login WHERE Username='%s'", username)
	_, err := db.Query(query)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func edituserKey(db *sql.DB, username, accesskey string) {
	query := fmt.Sprintf("UPDATE login_db.login SET AccessKey='%s' WHERE Username='%s'", accesskey, username)
	_, err := db.Query(query)
	if err != nil {
		fmt.Println(err)
		return
	}
}
