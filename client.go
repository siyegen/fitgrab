package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type CredGrabber interface {
	Login(username, password string) error
	Credentials() (*LoginData, error)
}

type FitocracyCredGrabber struct {
	Username  string
	SessionID string
	CSRFToken string
	UserID    int
}

func extractCookieValue(cookies []*http.Cookie, name string, outValue interface{}) error {
	for _, cookie := range cookies {
		if cookie.Name == name {
			outValue = cookie.Value
			return nil
		}
	}
	return fmt.Errorf("Couldn't find %s cookie", name)
}

func (f *FitocracyCredGrabber) Login(username, password string) error {
	// Get csrfToken
	client := &http.Client{}
	CSRFTokenResp, err := client.Get(tokenUrl)
	if err != nil {
		return err
	}

	var token string
	err = extractCookieValue(CSRFTokenResp.Cookies(), "csrftoken", &token)
	if err != nil {
		return err
	}

	// build form to login
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	form.Set("csrfmiddlewaretoken", token)
	form.Set("is_username", "1")
	form.Set("json", "1")

	// set csrftoken, content-type, and referer
	req, err := http.NewRequest("POST", loginUrl, strings.NewReader(form.Encode()))
	req.AddCookie(&http.Cookie{
		Name:  "csrftoken",
		Value: token,
	})
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Referer", "https://www.fitocracy.com") // Required

	// Post to loginUrl
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Bad statusCode: %d", resp.StatusCode)
	}
	var sessionID string
	err = extractCookieValue(resp.Cookies(), "sessionid", &sessionID)
	if err != nil {
		return err
	}
	fitUserID, err := strconv.Atoi(resp.Header["X-Fitocracy-User"][0])
	if err != nil {
		fmt.Println("Couldn't parse user id")
		return err
	}

	// Get sessionID from it, save it and csrf to uh, something
	f.UserID = fitUserID
	f.Username = username
	f.SessionID = sessionID
	f.CSRFToken = token

	return nil
}

func (f *FitocracyCredGrabber) Credentials() (*LoginData, error) {
	if f.Username == "" || f.SessionID == "" || f.CSRFToken == "" {
		return nil, fmt.Errorf("Invalid Credentials, call Login()")
	}
	return &LoginData{
		FitocracyUser: f.UserID,
		Username:      f.Username,
		SessionID:     f.SessionID,
		CSRFToken:     f.CSRFToken,
	}, nil
}

type FitGrabberClient struct {
	HTTPClient  *http.Client
	Credentials CredGrabber

	Username  string
	SessionID string
	CSRFToken string
	UserID    int
}

func NewFitGrabberClient(username, password string) (*FitGrabberClient, error) {
	credentials := &FitocracyCredGrabber{}
	err := credentials.Login(username, password)
	if err != nil {
		return nil, err
	}

	return &FitGrabberClient{
		HTTPClient:  &http.Client{},
		Credentials: credentials,
	}, nil
}
