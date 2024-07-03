package main

import (
	"os"
	"time"
)

func WriteTimeToFile() {

	currentTime := time.Now()

	// Format the time as a string
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	// Open a file for writing. If the file doesn't exist, create it, or append to the file
	file, err := os.OpenFile("time.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	// Write the time string to the file
	_, err = file.WriteString(formattedTime + "\r\n")
	if err != nil {
		return
	}

}
