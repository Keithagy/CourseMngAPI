package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type Course struct {
	Code string
	Info CourseInfo
}

type CourseInfo struct {
	Title      string `json:"Title"`
	Instructor string `json:"Instructor"`
	School     string `json:"School"`
}

type Account struct {
	Username string
	Password string
}

const (
	baseURL  = "http://localhost:5000/CMS/v1/courses"
	loginURL = "http://localhost:2000/users/v1"
	header   = "\n------------------------------\nCourse Management Service\n------------------------------\n"

	passwordHeader  = "TEAG*herd9tank-twis"
	accessKeyHeader = "keay*kak3jegh.BOB"
	usernameHeader  = "fer_ROUX9bam!preb"
)

var (
	accessKey string
	// Error Handling
	// errInvalid signals that a disallowed blank input was provided
	errInvalid = errors.New("ERROR: INVALID INPUT, TRY AGAIN")
	// errTerminate signals an exit from a menu
	errTerminate = errors.New("EXITING")
)

func main() {
	running := true
	for running {
		input, selection, err := menuSelection(
			"Login to Existing Account",
			"Create New Account")

		switch err {
		case nil:
			fmt.Println("Selected [", input, "] : ", selection)
			fmt.Printf("\n")
			switch input {
			case 0: // Login to Existing Account
				fmt.Println("Login")
				fmt.Println("Please input your username: ")
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				username := scanner.Text()

				fmt.Println("Please input your password: ")
				scanner = bufio.NewScanner(os.Stdin)
				scanner.Scan()
				password := scanner.Text()

				pwCookie := &http.Cookie{
					Name:  passwordHeader,
					Value: password,
				}
				request, err := http.NewRequest(http.MethodGet, loginURL+"/"+username+"/false", bytes.NewBuffer(nil))
				if err != nil {
					fmt.Printf("The HTTP request creation failed with error %s\n", err)
				} else {
					request.AddCookie(pwCookie)
					client := &http.Client{}
					response, err := client.Do(request)
					if err != nil {
						fmt.Printf("The HTTP request execution failed with error %s\n", err)
					} else {
						cookies := response.Cookies()
						failed := true
						for _, cookie := range cookies {
							if cookie.Name == accessKeyHeader {
								accessKey = cookie.Value
								failed = false
								fmt.Println("Welcome,", username)
								continue
							}
						}
						if failed {
							fmt.Println("Invalid username and/or password.")
						}
					}
				}

			case 1: // Create New Account
				fmt.Println("New Account")
				fmt.Println("Please input your username: ")
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				username := scanner.Text()

				fmt.Println("Please input your password: ")
				scanner = bufio.NewScanner(os.Stdin)
				scanner.Scan()
				password := scanner.Text()

				account := Account{username, password}
				jsonData, _ := json.Marshal(account)

				request, err := http.NewRequest(http.MethodPost, loginURL+"/"+username+"/false", bytes.NewBuffer(jsonData))
				if err != nil {
					fmt.Printf("The HTTP request creation failed with error %s\n", err)
				} else {
					request.Header.Set("Content-Type", "application/json")
					client := &http.Client{}
					response, err := client.Do(request)
					if err != nil {
						fmt.Printf("The HTTP request execution failed with error %s\n", err)
					} else {
						data, _ := ioutil.ReadAll(response.Body)
						fmt.Println(response.StatusCode)
						fmt.Println(string(data))
						response.Body.Close()
					}
				}
			}
		case errTerminate:
			fmt.Println(err)
			running = false
		}

		for validated := validate(accessKey); validated; validated = validate(accessKey) {
			input, selection, err := menuSelection(
				"View All Courses",
				"Access Specific Course",
				"Filter Courses by Criteria",
				"Add New Course",
				"Edit Existing Course",
				"Delete Existing Course")

			switch err {
			case nil:
				fmt.Println("Selected [", input, "] : ", selection)
				fmt.Printf("\n")
				switch input {
				case 0:
					// View All Courses
					getCourse("")
				case 1:
					// Access Specific Course
					var code string
					fmt.Println("Please input an EXISTING course code: ")
					scanner := bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						fmt.Printf("You inputted %s\n", scanner.Text())
					}
					code = scanner.Text()

					getCourse(code)
				case 2:
					// Filter Courses by Criteria
					filterCourses := true
					for filterCourses {
						input, selection, err := menuSelection(
							"Filter by Course Title",
							"Filter by Course Instructor",
							"Filter by Course Faculty")

						switch err {
						case nil:
							fmt.Println("Selected [", input, "] : ", selection)
							fmt.Printf("%s\n", header)
							var keyword string
							switch input {
							case 0:
								// Filter by Course Title
								fmt.Println("Please input title search term (SUBSTRING SEARCH, case-sensitive): ")
								scanner := bufio.NewScanner(os.Stdin)
								if scanner.Scan() {
									fmt.Printf("You inputted %s\n", scanner.Text())
								}
								keyword = scanner.Text()

								searchCourse("Title", keyword)
							case 1:
								// Filter by Course Instructor
								fmt.Println("Please input Instructor search term (SUBSTRING SEARCH, case-sensitive): ")
								scanner := bufio.NewScanner(os.Stdin)
								if scanner.Scan() {
									fmt.Printf("You inputted %s\n", scanner.Text())
								}
								keyword = scanner.Text()

								searchCourse("Instructor", keyword)
							case 2:
								// Filter by Course Faculty
								fmt.Println("Please input Facutly search term (SUBSTRING SEARCH, case-sensitive): ")
								scanner := bufio.NewScanner(os.Stdin)
								if scanner.Scan() {
									fmt.Printf("You inputted %s\n", scanner.Text())
								}
								keyword = scanner.Text()

								searchCourse("School", keyword)
							}
						case errTerminate:
							fmt.Println(err)
							filterCourses = false
						}
					}
				case 3:
					// Add New Course
					getCourse("")

					var code string
					var course CourseInfo

					fmt.Println("Please input a UNIQUE course code: ")
					scanner := bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						fmt.Printf("You inputted %s\n", scanner.Text())
					}
					code = scanner.Text()

					fmt.Println("Please input the course title: ")
					scanner = bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						fmt.Printf("You inputted %s\n", scanner.Text())
					}
					course.Title = scanner.Text()

					fmt.Println("\nPlease input the course instructor: ")
					scanner = bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						fmt.Printf("You inputted %s\n", scanner.Text())
					}
					course.Instructor = scanner.Text()

					fmt.Println("\nPlease input the course school: ")
					scanner = bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						fmt.Printf("You inputted %s\n", scanner.Text())
					}
					course.School = scanner.Text()

					addCourse(code, course)
				case 4:
					// Edit Existing Course
					getCourse("")

					var code string
					fmt.Println("Please input an EXISTING course code: ")
					scanner := bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						fmt.Printf("You inputted %s\n", scanner.Text())
					}
					code = scanner.Text()

					toEdit := (getCourse(code))[code]
					if toEdit.Title == "" {
						fmt.Println("Course code inputted does not exist!")
						continue
					}

					fmt.Println("Please update the course title (input blank if no change): ")
					scanner = bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						fmt.Printf("You inputted %s\n", scanner.Text())
					}
					if scanner.Text() != "" {
						toEdit.Title = scanner.Text()
					}

					fmt.Println("\nPlease update the course instructor (input blank if no change): ")
					scanner = bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						fmt.Printf("You inputted %s\n", scanner.Text())
					}
					if scanner.Text() != "" {
						toEdit.Instructor = scanner.Text()
					}

					fmt.Println("\nPlease update the course school (input blank if no change): ")
					scanner = bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						fmt.Printf("You inputted %s\n", scanner.Text())
					}
					if scanner.Text() != "" {
						toEdit.School = scanner.Text()
					}

					updateCourse(code, toEdit)
				case 5:
					// Delete Existing Course
					getCourse("")

					var code string
					fmt.Println("Please input an EXISTING course code: ")
					scanner := bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						fmt.Printf("You inputted %s\n", scanner.Text())
					}
					code = scanner.Text()

					deleteCourse(code)
				}
			case errTerminate:
				fmt.Println(err)
				accessKey = ""
			}
		}
	}
}

// Utility Functions
func menuSelection(options ...string) (int, string, error) {
	optioncount := len(options)
	fmt.Print(header)
	printOptions(options)
	input, err := menuChoice(optioncount)
	if err == errTerminate {
		return -9, "", err
	}
	for err != nil {
		fmt.Printf("%v\n\n", err)
		fmt.Print(header)
		printOptions(options)
		input, err = menuChoice(optioncount)
	}
	selection := options[input]
	return input, selection, err
}

func printOptions(options []string) {
	if len(options) == 0 {
		fmt.Println("No options currently available!")
	} else {
		for index := range options {
			fmt.Println("[", index, "] :", options[index])
		}
	}
}

func menuChoice(optioncount int) (int, error) {

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Please input your option (0 to %d)\n", optioncount-1)
	fmt.Println("Enter -9 to exit/terminate.")
	scanner.Scan()

	input, err := strconv.Atoi(scanner.Text())

	if err != nil {
		return -1, errInvalid
	}

	if input == -9 {
		// Exit Option selected
		return input, errTerminate
	} else if input >= 0 && input <= optioncount-1 {
		// Valid Input
		return input, nil
	} else {
		// Invalid Input
		return input, errInvalid
	}
}

// CRUD Functions
func getCourse(code string) map[string]CourseInfo {
	url := baseURL
	if code != "" {
		url = baseURL + "/" + code + "?" + accessKeyHeader + "=" + accessKey
	} else {
		url = baseURL + "?" + accessKeyHeader + "=" + accessKey
	}
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return nil
	} else {
		defer response.Body.Close()
		raw, _ := ioutil.ReadAll(response.Body)
		var data map[string]CourseInfo
		json.Unmarshal(raw, &data)
		if len(data) == 0 {
			fmt.Println("No courses found.")
			return nil
		} else {
			for code, course := range data {
				fmt.Println(code, "-", course.Title, "|| Taught by:", course.Instructor, "|| Faculty:", course.School)
			}
			return data
		}
	}
}

func addCourse(code string, jsonData CourseInfo) {
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post(baseURL+"/"+code+"?"+accessKeyHeader+"="+accessKey, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		defer response.Body.Close()
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(response.StatusCode)
		fmt.Println(string(data))
	}
}

func updateCourse(code string, jsonData CourseInfo) {
	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest(http.MethodPut, baseURL+"/"+code+"?"+accessKeyHeader+"="+accessKey, bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		request.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			fmt.Println(response.StatusCode)
			fmt.Println(string(data))
			response.Body.Close()
		}
	}
}

func deleteCourse(code string) {
	request, err := http.NewRequest(http.MethodDelete, baseURL+"/"+code+"?"+accessKeyHeader+"="+accessKey, nil)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			fmt.Println(response.StatusCode)
			fmt.Println(string(data))
			response.Body.Close()
		}
	}
}

func searchCourse(criteria string, keyword string) {
	url := baseURL + "?" + accessKeyHeader + "=" + accessKey + "&" + criteria + "=" + keyword
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		defer response.Body.Close()
		raw, _ := ioutil.ReadAll(response.Body)
		var data map[string]CourseInfo
		json.Unmarshal(raw, &data)
		if len(data) == 0 {
			fmt.Println("No courses found.")
		} else {
			for code, course := range data {
				fmt.Println(code, "-", course.Title, "|| Taught by:", course.Instructor, "|| Faculty:", course.School)
			}
		}
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
			fmt.Println("Access Key confirmed.")
		} else {
			fmt.Println("No valid Access Key found.")
		}
		return result.Validated
	}
}
