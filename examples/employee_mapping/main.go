// This a app used to test the implementation and dev of fieldmap
package main

import (
	"fmt"
	"time"

	"github.com/christopher-eglinton/fieldmap"
)

type Employee struct {
	FirstName string
	LastName  string
	Email     string
	Active    bool
	StartDate time.Time
}

func main() {
	input := map[string]any{
		"employee": map[string]any{
			"givenName":  "Chris",
			"familyName": "Eglinton",
		},
		"workEmail": " CHRIS@EXAMPLE.COM ",
		"employment": map[string]any{
			"active":    "true",
			"startDate": "2024-01-15",
		},
	}

	cfg := fieldmap.Config{
		Rules: []fieldmap.Rule{
			{From: "employee.givenName", To: "FirstName", Required: true},
			{From: "employee.familyName", To: "LastName", Required: true},
			{From: "workEmail", To: "Email", Required: true, Transform: fieldmap.TrimLower()},
			{From: "employment.active", To: "Active", Transform: fieldmap.StringToBool()},
			{From: "employment.startDate", To: "StartDate", Transform: fieldmap.ParseTime("2006-01-02")},
		},
	}

	var emp Employee
	if err := fieldmap.Apply(cfg, input, &emp); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", emp)
}
