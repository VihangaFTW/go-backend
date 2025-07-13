package validator

import (
	"fmt"
	"net/mail"
	"regexp"
	"unicode/utf8"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString
	isValidFullName = regexp.MustCompile(`^[a-zA-Z\\s]+$`).MatchString
)

func ValidateString(value string, minLength int, maxLength int) error {

	n := utf8.RuneCountInString(value)

	if n < minLength || n > maxLength {
		return fmt.Errorf("must contain from %d to %d characters", minLength, maxLength)
	}

	return nil

}

func ValidateUsername(value string) error {

	if err := ValidateString(value, 6, 30); err != nil {
		return err
	}

	if !isValidUsername(value) {
		return fmt.Errorf("username must contain only letters, digits or underscore")
	}

	return nil
}

func ValidatePassword(password string) error {
	// password must contain at least 6 characters
	// no restrictions on password characters
	return ValidateString(password, 6, 100)
}

func ValidateEmail(email string) error {
	if err := ValidateString(email, 3, 200); err != nil {
		return err
	}

	// check if email is valid according to official standards
	_, err := mail.ParseAddress(email)

	if err != nil {
		return fmt.Errorf("email address is invalid")
	}

	return nil
}

func ValidateFullName(value string) error {

	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}

	if !isValidFullName(value) {
		return fmt.Errorf("must contain only lowercase and uppercase english letters")
	}

	return nil
}


