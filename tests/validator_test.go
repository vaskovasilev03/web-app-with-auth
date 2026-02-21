package tests

import (
	"testing"
	"web-app/internal/validator"
)

func TestIsValidEmail(t *testing.T) {
	emailTests := []struct {
		email   string
		isValid bool
	}{
		{"test@example.com", true},
		{"This is not an email", false},
		{"mynewemail@domain.bg.com", true},
		{"twodots..@example.com", false},
		{"endingwithdot.@example.com", false},
		{".startingwtihdot@example.com", false},
		{"", false},
		{"   ", false},
		{"user@.com", false},
	}

	for _, testData := range emailTests {
		if validator.IsValidEmail(testData.email) != testData.isValid {
			t.Errorf("IsValidEmail(%q) = %v; want %v", testData.email, !testData.isValid, testData.isValid)
		}
	}
}

func TestIsValidPassword(t *testing.T) {
	passwordTests := []struct {
		password string
		isValid  bool
	}{
		{"Password123!", true},
		{"short1!", false},
		{"nouppercase1!", false},
		{"NOLOWERCASE1!", false},
		{"NoSpecialChar1", false},
		{"   ", false},
	}

	for _, testData := range passwordTests {
		if validator.IsValidPassword(testData.password) != testData.isValid {
			t.Errorf("IsValidPassword(%q) = %v; want %v", testData.password, !testData.isValid, testData.isValid)
		}
	}
}

func TestIsValidName(t *testing.T) {
	nameTests := []struct {
		name    string
		isValid bool
	}{
		{"John Doe", true},
		{"", false},
		{"   ", false},
		{"A very long name that exceeds the maximum length of fifty characters", false},
		{"NameWith123", false},
		{"Name-With-Dash", false},
	}

	for _, testData := range nameTests {
		if validator.IsValidName(testData.name) != testData.isValid {
			t.Errorf("IsValidName(%q) = %v; want %v", testData.name, !testData.isValid, testData.isValid)
		}
	}
}
