package fitgrabber

import (
	"fmt"
	"io/ioutil"
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

var loginUrl = "https://www.fitocracy.com/accounts/login/"
var tokenUrl = "https://www.fitocracy.com/"

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

func extractCookieValue(cookies []*http.Cookie, name string) (string, error) {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value, nil
		}
	}
	return "", fmt.Errorf("Couldn't find %s cookie", name)
}

func (f *FitocracyCredGrabber) Login(username, password string) error {
	// Get csrfToken
	client := &http.Client{}
	CSRFTokenResp, err := client.Get(tokenUrl)
	if err != nil {
		return err
	}

	token, err := extractCookieValue(CSRFTokenResp.Cookies(), "csrftoken")
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

	fmt.Printf("Req with csrf: %s\n\n", token)
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
	sessionID, err := extractCookieValue(resp.Cookies(), "sessionid")
	if err != nil {
		return err
	}
	fitUserID, err := strconv.Atoi(resp.Header["X-Fitocracy-User"][0])
	if err != nil {
		fmt.Println("Couldn't parse user id")
		return err
	}

	fmt.Println("everything good?", resp.StatusCode)
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

type FitocracyClient struct {
	HTTPClient  *http.Client
	Credentials CredGrabber

	Username  string
	SessionID string
	CSRFToken string
	UserID    int
}

var (
	activityList = "https://www.fitocracy.com/get_user_activities/"
	activity     = "https://www.fitocracy.com/_get_activity_history_json/?activity-id="
)

func (f *FitocracyClient) GetActivity(id int) {
	credentials, _ := f.Credentials.Credentials()
	url := fmt.Sprintf(activity+"%d", id)
	fmt.Println("url", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// return err
		return
	}
	req.Header.Add("Referer", "https://www.fitocracy.com") // Required
	req.AddCookie(&http.Cookie{
		Name:  "csrfmiddlewaretoken",
		Value: credentials.CSRFToken,
	})
	req.AddCookie(&http.Cookie{
		Name:  "sessionid",
		Value: credentials.SessionID,
	})

	resp, err := f.HTTPClient.Do(req)
	if err != nil {
		fmt.Println("Error!", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error!", err)
		return
	}
	if resp.StatusCode == 200 {
		fmt.Println("Body\n\n", string(body))
	}
}

func (f *FitocracyClient) GetActivityList() (string, error) {
	credentials, _ := f.Credentials.Credentials()
	url := fmt.Sprintf(activityList+"%d/", credentials.FitocracyUser)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Referer", "https://www.fitocracy.com") // Required
	req.AddCookie(&http.Cookie{
		Name:  "csrfmiddlewaretoken",
		Value: credentials.CSRFToken,
	})
	req.AddCookie(&http.Cookie{
		Name:  "sessionid",
		Value: credentials.SessionID,
	})

	resp, err := f.HTTPClient.Do(req)
	if err != nil {
		fmt.Println("Error!", err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error!", err)
		return "", err
	}
	fmt.Println("Body\n\n", string(body))
	return string(body), nil
}

func NewFitocracyClient(username, password string) (*FitocracyClient, error) {
	credentials := &FitocracyCredGrabber{}
	err := credentials.Login(username, password)
	if err != nil {
		return nil, err
	}

	return &FitocracyClient{
		HTTPClient:  &http.Client{},
		Credentials: credentials,
	}, nil
}
