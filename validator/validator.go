package validator

import (
	"regexp"
)

const (
	alphaNum  = "^[a-zA-Z0-9_]*$"
	alpha     = "^[a-zA-Z_]*$"
	mobileNum = "^[8-9][0-9]{7}$"
)

func IsEmpty(input string) bool {
	return len(input) == 0
}

func IsAlphabet(input string) bool {
	regex := regexp.MustCompile(alpha)
	return regex.MatchString(input)
}

func IsAlphaNumeric(input string) bool {
	regex := regexp.MustCompile(alphaNum)
	return regex.MatchString(input)
}

func IsMobileNumber(input string) bool {
	regex := regexp.MustCompile(mobileNum)
	return regex.MatchString(input)
}
