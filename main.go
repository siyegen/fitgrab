package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type LoginData struct {
	FitocracyUser int
	Username      string
	SessionID     string
	CSRFToken     string
}

var testUrl = "https://www.fitocracy.com/profile/siyegen/?feed"
var testError = errors.New("net/http: use last response")

var loginUrl = "https://www.fitocracy.com/accounts/login/"
var tokenUrl = "https://www.fitocracy.com/"

var testClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return testError
	},
}

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

	statusCode, err := testCall("", "")
	if err != nil {
		fmt.Println("Bad response", err)
	}
	fmt.Println("Status", statusCode)
	// login logic
	loginData, err := login(username, password)
	if err != nil {
		fmt.Println("Bad login", err)
	}

	fmt.Printf("Valid login data: %+v\n", loginData)
	// call again and see different status
	statusCode, err = testCall(loginData.SessionID, loginData.CSRFToken)
	if err != nil {
		fmt.Println("Bad response", err)
	}
	fmt.Println("Status", statusCode)
}

func login(user, password string) (*LoginData, error) {
	client := &http.Client{}
	var token string
	// get csrfmiddlewaretoken
	csrfTokenResponse, err := client.Get(tokenUrl)
	if err != nil {
		fmt.Println("Err getting token")
		return nil, err
	}
	for _, cookie := range csrfTokenResponse.Cookies() {
		if cookie.Name == "csrftoken" {
			token = cookie.Value
		}
	}
	if token == "" {
		return nil, fmt.Errorf("No csrftoken")
	}
	fmt.Printf("Req with csrf: %s\n\n", token)

	form := url.Values{}
	form.Add("username", user)
	form.Add("password", password)
	form.Add("csrfmiddlewaretoken", token)
	form.Add("is_username", "1")
	form.Add("json", "1")

	req, err := http.NewRequest("POST", loginUrl, strings.NewReader(form.Encode()))
	req.AddCookie(&http.Cookie{
		Name:  "csrftoken",
		Value: token,
	})
	if err != nil {
		fmt.Println("Err building request")
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// req.Header.Add("Host", "www.fitocracy.com")
	// req.Header.Add("Origin", "https://www.fitocracy.com")
	req.Header.Add("Referer", "https://www.fitocracy.com") // Required
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Err sending post")
		return nil, err
	}
	defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// fmt.Println("resp", body)
	var loggedInSessionID string
	for k, v := range resp.Header {
		fmt.Printf("%s: %s\n", k, v)
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "sessionid" {
			loggedInSessionID = cookie.Value
		}
	}
	fmt.Println("login status", resp.StatusCode)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad return code: %d", resp.StatusCode)
	}
	fitUserID, err := strconv.Atoi(resp.Header["X-Fitocracy-User"][0])
	if err != nil {
		fmt.Println("Couldn't parse user id")
		return nil, err
	}

	loginData := &LoginData{
		FitocracyUser: fitUserID,
		Username:      user,
		SessionID:     loggedInSessionID,
		CSRFToken:     token,
	}
	return loginData, nil
}

func testCall(sessionID, csrfToken string) (int, error) {
	req, err := http.NewRequest("GET", testUrl, nil)
	if err != nil {
		return 0, nil
	}

	if sessionID != "" {
		req.AddCookie(&http.Cookie{
			Name:  "sessionid",
			Value: sessionID,
		})
	}
	if csrfToken != "" {
		req.AddCookie(&http.Cookie{
			Name:  "csrfmiddlewaretoken",
			Value: csrfToken,
		})
	}
	req.Header.Add("Referer", "https://www.fitocracy.com") // Required

	resp, err := testClient.Do(req)
	if err != nil {
		if strings.HasSuffix(err.Error(), testError.Error()) {
			fmt.Println("Not following redirect")
		} else {
			return resp.StatusCode, err
		}
	}

	return resp.StatusCode, nil
}
