package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

var logger *log.Logger = log.New(os.Stdout, "FitGrabber", log.LstdFlags)

type Options func(*FitGrabber) error

func FileLocation(location string) Options {
	return func(f *FitGrabber) error {
		return f.setLocation(location)
	}
}

func CallDelay(delay time.Duration) Options {
	return func(f *FitGrabber) error {
		return f.setDelay(delay)
	}
}

var errMissingUsernamePassword = errors.New("missing username/password")

type FitGrabber struct {
	client       *FitocracyClient
	fileLocation string
	callDelay    time.Duration
}

func (f *FitGrabber) setDelay(delay time.Duration) error {
	if delay > time.Second*30 {
		logger.Printf("delay %d too long, setting to max of 30", delay)
		delay = time.Second * 30
	}
	f.callDelay = delay
	return nil
}

func (f *FitGrabber) setLocation(location string) error {
	absLocation, err := filepath.Abs(location)
	if err != nil {
		return err
	}
	f.fileLocation = absLocation
	return nil
}

func NewFitGrabber(username, password string, options ...Options) (*FitGrabber, error) {
	if username == "" || password == "" {
		logger.Println("Error: Username / Password required")
		return nil, errMissingUsernamePassword
	}

	fitClient, err := NewFitocracyClient(username, password)
	if err != nil {
		return nil, err
	}
	fit := &FitGrabber{
		client:       fitClient,
		fileLocation: ".fitgrabber",
		callDelay:    time.Second * 3,
	}

	for _, opts := range options {
		err := opts(fit)
		if err != nil {
			return nil, err
		}
	}

	err = os.MkdirAll(fit.fileLocation, 0777)
	if err != nil {
		return nil, err
	}
	return fit, nil
}

func main() {
	fmt.Println("Fitocracy stats grabber")
	// Option handling
	var username, password string
	flag.StringVar(&username, "username", "", "Fitocracy username")
	flag.StringVar(&password, "password", "", "Fitocracy password")
	flag.Parse()

	// File location handling
	osUser, err := user.Current()
	if err != nil {
		logger.Fatal("Couldn't get current user")
	}
	exerciseLocation := filepath.Join(osUser.HomeDir, ".fitgrabber")

	fitGrabber, err := NewFitGrabber(username, password,
		CallDelay(time.Second*0),
		FileLocation(exerciseLocation),
	)
	if err != nil {
		logger.Fatal("Error creating command", err)
	}

	now := time.Now()
	fileName := fmt.Sprintf("%s_activities", now.Format("2006_01_02_15_04"))
	f, err := os.Create(filepath.Join(fitGrabber.fileLocation, fileName))

	// actually call the endpoint
	allLifts, _ := fitGrabber.client.GetActivityList()
	if err != nil {
		flag.Usage()
		logger.Fatal("Error writing file", err)
	}

	// Store the results
	defer f.Close()
	f.WriteString(allLifts)
	fmt.Println("==================================\n\n")

	// fitClient.GetActivity(174)

}
