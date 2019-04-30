package middleware

import "testing"

func TestValidateEmail(t *testing.T) {
	invalidEmails := []string{
		"test",
		"@example.com",
		"",
		"test@example",
		"test.com",
		"test@e@e.com",
		"test.com@e",
		"test@e.123",
	}
	validEmails := []string{
		"test@example.com",
		"test.test@e.com",
		"TEST@e.au",
		"1234@e.in",
	}
	for _, em := range invalidEmails {
		got := validEmail(em)
		if got {
			t.Errorf("Email validate failed. Expected=false but received=%v", got)
		}
	}
	for _, em := range validEmails {
		got := validEmail(em)
		if !got {
			t.Errorf("Email validate failed. Expected=true but received=%v", got)
		}
	}
}

func TestValidatePhone(t *testing.T) {
	invalidPhoneNumbers := []string{
		"",
		"0",
		"1000000000",
		"6000000000",
		"90000000000",
		"900000000",
	}

	validPhoneNumbers := []string{
		"9000000000",
		"8000000000",
		"7000000000",
	}

	for _, p := range invalidPhoneNumbers {
		got := validPhoneNumber(p)
		if got {
			t.Errorf("Phone validate failed. Expected=false but received=%v", got)
		}
	}

	for _, p := range validPhoneNumbers {
		got := validPhoneNumber(p)
		if !got {
			t.Errorf("Phone validate failed. Expected=true but received=%v", got)
		}
	}
}
