package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const accessKeyHeader = "keay*kak3jegh.BOB"

type CourseInfo struct {
	Title      string `json:"Title"`
	Instructor string `json:"Instructor"`
	School     string `json:"School"`
}

// Storing courses in-memory on the REST API
var (
	courses map[string]CourseInfo // map keys are string course codes

	db  *sql.DB
	err error // for sql.Open in init()
)

func init() {
	// connect to DB for courses
	db, err = sql.Open("mysql", "root:veg-kluh!PRIW3hirt@tcp(127.0.0.1:3306)/courses_db")
	if err != nil {
		panic(err.Error)
	} else {
		fmt.Println("Courses database opened successfully.")
		// populate courses with data from DB, if any
		courses = getRecords(db)
	}
}

func main() {
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/CMS/v1/", home)
	router.HandleFunc("/CMS/v1/courses", allcourses)
	router.HandleFunc("/CMS/v1/courses/{courseid}", course).Methods("GET", "PUT", "POST", "DELETE")

	fmt.Println("Listening at port 5000")
	log.Fatal(http.ListenAndServe(":5000", router))
}

// Handler Functions
func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Course Management Service, please use the client functions to navigate the application.")
}

func allcourses(w http.ResponseWriter, r *http.Request) {
	kv := r.URL.Query()
	key := kv[accessKeyHeader][0]
	if !validate(key) {
		return
	}
	// returns the key/value pairs in the query string as a map object

	if len(kv) > 1 {
		subsetted := make(map[string]CourseInfo)
		for k, v := range kv {
			for code, course := range courses {
				switch k {
				case "Title":
					if strings.Contains(course.Title, v[0]) {
						subsetted[code] = course
					}
				case "Instructor":
					if strings.Contains(course.Instructor, v[0]) {
						subsetted[code] = course
					}
				case "School":
					if strings.Contains(course.School, v[0]) {
						subsetted[code] = course
					}
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(subsetted)
	} else { // No search criteria received, return all courses
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(courses)
	}
}

func course(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	key := v[accessKeyHeader][0]
	if !validate(key) {
		return
	}

	params := mux.Vars(r)
	fmt.Println("Params:", params)

	if r.Method == "GET" {
		fmt.Println("Course view called")
		if _, ok := courses[params["courseid"]]; ok {
			json.NewEncoder(w).Encode(map[string]CourseInfo{params["courseid"]: courses[params["courseid"]]})
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - No course found"))
			return
		}
	}

	if r.Method == "DELETE" {
		fmt.Println("Course delete called")
		if _, ok := courses[params["courseid"]]; ok {
			delete(courses, params["courseid"])
			deleteRecord(db, params["courseid"])
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("202 - Course deleted: " + params["courseid"]))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - No course found"))
			return
		}
	}

	if r.Header.Get("Content-type") == "application/json" {
		// POST is for creating new course
		if r.Method == "POST" {
			fmt.Println("Course add called")
			// read the string sent to the service
			var newCourse CourseInfo
			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}

			// convert JSON to object
			json.Unmarshal(reqBody, &newCourse)

			if newCourse.Title == "" || newCourse.Instructor == "" || newCourse.School == "" {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Blank fields detected. Course not added successfully."))
				return
			}

			// check if course exists; add only if course does not exist
			if _, ok := courses[params["courseid"]]; !ok {
				courses[params["courseid"]] = newCourse
				insertRecord(db, params["courseid"], newCourse.Title, newCourse.Instructor, newCourse.School)
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("201 - Course added : " + params["courseid"]))
			} else {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Duplicate entry detected."))
			}
		}

		// PUT is for creating or updating existing courses
		if r.Method == "PUT" {
			fmt.Println("Course update called")
			var newCourse CourseInfo
			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}

			json.Unmarshal(reqBody, &newCourse)

			if newCourse.Title == "" {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Please supply course information in JSON format"))
				return
			}

			// check if course exists; create course if it does not and update course if it does
			if _, ok := courses[params["courseid"]]; !ok {
				courses[params["courseid"]] = newCourse
				insertRecord(db, params["courseid"], newCourse.Title, newCourse.Instructor, newCourse.School)
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("201 - Course added: " + params["courseid"]))
			} else {
				// update course
				courses[params["courseid"]] = newCourse
				editRecord(db, params["courseid"], newCourse.Title, newCourse.Instructor, newCourse.School)
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte("202 - Course updated: " + params["courseid"]))
			}
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("422 - Please supply course information in JSON format"))
			return
		}
	}
}

// Database operations

func getRecords(db *sql.DB) map[string]CourseInfo {
	result := make(map[string]CourseInfo)

	existing, err := db.Query("Select * FROM courses_db.courses")
	if err != nil {
		panic(err.Error())
	}
	for existing.Next() {
		var code string
		var course CourseInfo

		err = existing.Scan(&code, &course.Title, &course.Instructor, &course.School)
		if err != nil {
			panic(err.Error())
		}
		result[code] = course
		fmt.Println("Course", code, "found from DB and loaded into memory.")
	}
	return result
}

func deleteRecord(db *sql.DB, code string) {
	query := fmt.Sprintf("DELETE FROM courses WHERE ID='%s'", code)
	_, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
}

func insertRecord(db *sql.DB, code, title, instructor, faculty string) {
	query := fmt.Sprintf("INSERT INTO courses VALUES ('%s', '%s', '%s', '%s')", code, title, instructor, faculty)
	_, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
}

func editRecord(db *sql.DB, code, title, instructor, faculty string) {
	query := fmt.Sprintf("UPDATE courses SET Title='%s', Instructor='%s', Faculty='%s' WHERE ID='%s'", title, instructor, faculty, code)
	_, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
}

func validate(accessKey string) bool {
	if accessKey == "nil" {
		fmt.Println("No valid Access Key found.")
		return false
	}
	validationURL := "http://localhost:2000/keys/v1/" + accessKey
	response, err := http.Get(validationURL)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return false
	} else {
		defer response.Body.Close()
		raw, _ := ioutil.ReadAll(response.Body)
		var result struct {
			Validated bool
		}
		json.Unmarshal(raw, &result)
		if result.Validated {
			fmt.Println("Access Key confirmed:", accessKey)
		} else {
			fmt.Println("No valid Access Key found.")
		}
		return result.Validated
	}
}
