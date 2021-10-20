package handler

import (
	"fmt"
	"strings"
	"time"
)

const (
	layoutISO = "2006-01-02"
)

func Time() (string, string, string) {

	now := time.Now()
	y, m, d := now.Date()

	dt := fmt.Sprintf("%d-%d-%d", y, int(m), d)
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		fmt.Println(err)
	}
	tm := now.In(location).Format("03:04pm")
	u := time.Now().In(location)
	zn := u.Format("MST")
	if strings.ContainsAny(zn, "PDT") {
		zn := "PT"
		return tm, zn, dt
	}
	return tm, zn, dt

}

//todo need to find out to pass time in AM and PM format.
