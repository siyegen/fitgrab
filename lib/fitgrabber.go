package fitgrabber

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"
)

var logger *log.Logger = log.New(os.Stdout, "FitGrabber", log.LstdFlags)

type Options func(*FitGrabber) error

func SaveLocation(location string) Options {
	return func(f *FitGrabber) error {
		return f.setSaveLocation(location)
	}
}

func StoreLocation(location string) Options {
	return func(f *FitGrabber) error {
		return f.setStoreLocation(location)
	}
}

func CallDelay(delay time.Duration) Options {
	return func(f *FitGrabber) error {
		return f.setDelay(delay)
	}
}

var errMissingUsernamePassword = errors.New("missing username/password")

type FitGrabber struct {
	Client        *FitocracyClient
	SaveLocation  string // For saving the exercises/numbers
	StoreLocation string // For saving map of accessible data to use for calls
	CallDelay     time.Duration

	Logger *log.Logger
}

func (f *FitGrabber) setDelay(delay time.Duration) error {
	if delay > time.Second*30 {
		f.Logger.Printf("delay %d too long, setting to max of 30", delay)
		delay = time.Second * 30
	}
	f.CallDelay = delay
	return nil
}

func (f *FitGrabber) setSaveLocation(location string) error {
	absLocation, err := filepath.Abs(location)
	if err != nil {
		return err
	}
	f.SaveLocation = absLocation
	return nil
}

func (f *FitGrabber) setStoreLocation(location string) error {
	absLocation, err := filepath.Abs(location)
	if err != nil {
		return err
	}
	f.StoreLocation = absLocation
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
		Client:       fitClient,
		SaveLocation: ".fitgrabber",
		CallDelay:    time.Second * 3,
		Logger:       logger,
	}

	for _, opts := range options {
		err := opts(fit)
		if err != nil {
			return nil, err
		}
	}

	err = os.MkdirAll(fit.SaveLocation, 0777)
	if err != nil {
		return nil, err
	}
	return fit, nil
}
