package fieldmap

import (
	"testing"
	"time"
)

type testEmployee struct {
	FirstName string
	LastName  string
	Email     string
	Active    bool
	StartDate time.Time
}

func TestApply(t *testing.T) {
	input := map[string]any{
		"employee": map[string]any{
			"givenName":  "Tony",
			"familyName": "Soprano",
		},
		"workEmail": " TONE@DIMEO.COM ",
		"employment": map[string]any{
			"active":    "true",
			"startDate": "2026-03-28",
		},
	}

	cfg := Config{
		Rules: []Rule{
			{From: "employee.givenName", To: "FirstName", Required: true},
			{From: "employee.familyName", To: "LastName", Required: true},
			{From: "workEmail", To: "Email", Required: true, Transform: TrimLower()},
			{From: "employment.active", To: "Active", Transform: StringToBool()},
			{From: "employment.startDate", To: "StartDate", Transform: ParseTime("2006-01-02")},
		},
	}

	var emp testEmployee
	err := Apply(cfg, input, &emp)
	if err != nil {
		t.Fatalf("Apply returned unexpected error: %v", err)
	}

	if emp.FirstName != "Tony" {
		t.Fatalf("expected FirstName Tony, got %q", emp.FirstName)
	}
	if emp.LastName != "Soprano" {
		t.Fatalf("expected LastName Soprano, got %q", emp.LastName)
	}
	if emp.Email != "tone@dimeo.com" {
		t.Fatalf("expected Email tone@dimeo.com, got %q", emp.Email)
	}
	if !emp.Active {
		t.Fatalf("expected Active true, got false")
	}

	expectedDate, _ := time.Parse("2006-01-02", "2026-03-28")
	if !emp.StartDate.Equal(expectedDate) {
		t.Fatalf("expected StartDate %v, got %v", expectedDate, emp.StartDate)
	}
}
