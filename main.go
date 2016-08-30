package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	fitgrabber "github.com/siyegen/fitgrab/lib"
)

func main() {
	fmt.Println("Fitocracy stats grabber")
	// Option handling
	var username, password, saveFile string
	flag.StringVar(&username, "username", "", "Fitocracy username")
	flag.StringVar(&password, "password", "", "Fitocracy password")
	flag.StringVar(&saveFile, "file", "", "Save location")
	flag.Parse()

	// File location handling
	osUser, err := user.Current()
	if err != nil {
		log.Fatal("Couldn't get current user")
	}
	exerciseLocation := filepath.Join(osUser.HomeDir, ".fitgrabber")
	now := time.Now()
	fileName := filepath.Join(exerciseLocation, fmt.Sprintf("%s_activities", now.Format("2006_01_02_15_04")))

	fitGrabber, err := fitgrabber.NewFitGrabber(username, password,
		fitgrabber.CallDelay(time.Second*0),
		fitgrabber.SaveLocation(fileName),
	)
	if err != nil {
		log.Fatal("Error creating command", err)
	}

	f, err := os.Create(fitGrabber.SaveLocation)
	if err != nil {
		flag.Usage()
		fitGrabber.Logger.Fatal("Error writing file", err)
	}

	// actually call the endpoint
	allLifts, _ := fitGrabber.Client.GetActivityList()

	// Store the results
	defer f.Close()
	f.WriteString(allLifts)
	fmt.Println("==================================\n\n")

	// fitClient.GetActivity(174)

}
