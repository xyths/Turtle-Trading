package utils

import (
	"errors"
	"fmt"
	"log"
	"time"
)

const TimeLayout = "2006-01-02 15:04:05"

func ParseStartEndTime(start, end string) (startTime, endTime time.Time) {
	secondsEastOfUTC := int((8 * time.Hour).Seconds())
	beijing := time.FixedZone("Beijing Time", secondsEastOfUTC)

	startTime, err := time.ParseInLocation(TimeLayout, start, beijing)
	if err != nil {
		log.Fatalf("error start format: %s", start)
	}
	endTime, err = time.ParseInLocation(TimeLayout, end, beijing)
	if err != nil {
		log.Fatalf("error end format: %s", end)
	}
	if !startTime.Before(endTime) {
		err = errors.New(fmt.Sprintf("start time(%s) must before end time(%s)", startTime.String(), endTime.String()))
		log.Fatal(err)
	}
	return
}
