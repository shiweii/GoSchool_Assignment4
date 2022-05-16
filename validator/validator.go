package validator

import (
	"regexp"
	"unicode"
)

const (
	name              = "^[a-zA-Z_. ]*$"
	username          = "^[a-zA-Z0-9][a-zA-Z0-9\\_\\-\\.]*[a-zA-Z0-9]$"
	mobileNum         = "^[8-9][0-9]{7}$"
	usernameMinLength = 5
	usernameMaxLength = 20
	PasswordMinLength = 7
)

func IsEmpty(input string) bool {
	return len(input) == 0
}

func IsValidName(input string) bool {
	regex := regexp.MustCompile(name)
	return regex.MatchString(input)
}

// IsValidUsername
// Username consists max 20 characters
// Username consists of alphanumeric characters (a-zA-Z0-9), lowercase, or uppercase.
// Username allowed of the dot (.), underscore (_), and hyphen (-).
// The dot (.), underscore (_), or hyphen (-) must not be the first or last character.
func IsValidUsername(input string) bool {
	if len(input) < usernameMinLength || len(input) > usernameMaxLength {
		return false
	}
	regex := regexp.MustCompile(username)
	return regex.MatchString(input)
}

// IsValidPassword
// at least 1 number
// at least 1 upper case
// at least 1 special character
func IsValidPassword(input string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(input) >= PasswordMinLength {
		hasMinLen = true
	}
	for _, char := range input {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

func IsMobileNumber(input string) bool {
	regex := regexp.MustCompile(mobileNum)
	return regex.MatchString(input)
}
