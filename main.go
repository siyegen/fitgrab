package main

import (
	"flag"
	"fmt"
)

type LoginData struct {
	FitocracyUser int
	Username      string
	SessionID     string
	CSRFToken     string
}

var loginUrl = "https://www.fitocracy.com/accounts/login/"
var tokenUrl = "https://www.fitocracy.com/"

func main() {
	fmt.Println("Fitocracy stats grabber")
	var username, password string
	flag.StringVar(&username, "username", "", "Fitocracy username")
	flag.StringVar(&password, "password", "", "Fitocracy password")
	flag.Parse()

	if username == "" || password == "" {
		fmt.Println("Error: Username / Password required")
		flag.Usage()
		return
	}

	fitClient, err := NewFitGrabberClient(username, password)
	if err != nil {
		fmt.Println("damn, errors", err)
		return
	}
	creds, _ := fitClient.Credentials.Credentials()
	fmt.Printf("Creds %+v\n", creds)
	fitClient.GetActivityList()

}
