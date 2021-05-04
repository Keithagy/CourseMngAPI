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

	_ "github.com/go-sql-driver/mysql"
)

const (
	baseURL = "http://localhost:2000/users/v1"
	header  = "\n------------------------------\nAdmin Login Management\n------------------------------\n"
)

var (
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
			"Provision / Revoke Access Key",
			"Delete User Account")

		switch err {
		case nil:
			fmt.Println("Selected [", input, "] : ", selection)
			fmt.Printf("\n")
			response, err := http.Get(baseURL + "/")
			if err != nil {
				fmt.Printf("The HTTP request failed with error %s\n", err)
			} else {
				defer response.Body.Close()
				raw, _ := ioutil.ReadAll(response.Body)
				var data []string
				json.Unmarshal(raw, &data)
				if len(data) == 0 {
					fmt.Println("No users found.")
				} else {
					fmt.Println("Existing usernames:")
					for _, user := range data {
						fmt.Println(user)
					}
				}
			}
			switch input {
			case 0:
				fmt.Println("Please input the username to provide/revoke an access key: ")
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				username := scanner.Text()

				request, err := http.NewRequest(http.MethodPut, baseURL+"/"+username+"/true", bytes.NewBuffer(nil))
				if err != nil {
					fmt.Printf("The HTTP request creation failed with error %s\n", err)
				} else {
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
			case 1:
				fmt.Println("Please input the username of the account to delete: ")
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				username := scanner.Text()

				request, err := http.NewRequest(http.MethodDelete, baseURL+"/"+username+"/true", bytes.NewBuffer(nil))
				if err != nil {
					fmt.Printf("The HTTP request creation failed with error %s\n", err)
				} else {
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
