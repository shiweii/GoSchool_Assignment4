package validator

import (
	"testing"
)

func TestIsEmpty(t *testing.T) {
	got := IsEmpty("")
	res := true
	if got != res {
		t.Errorf("IsEmpty() = %t; want %t got %t", got, res, got)
	}

	got = IsEmpty("Hello World")
	res = false
	if got != res {
		t.Errorf("IsEmpty(Hello World) = %t; want %t got %t", got, res, got)
	}
}

func TestIsAlphabet(t *testing.T) {
	got := IsAlphabet("abcd")
	res := true
	if got != res {
		t.Errorf("IsAlphabet(abcd) = %t; want %t got %t", got, res, got)
	}

	got = IsAlphabet("Hello World")
	res = false
	if got != res {
		t.Errorf("IsAlphabet(Hello World) = %t; want %t got %t", got, res, got)
	}

	got = IsAlphabet("12345")
	res = false
	if got != res {
		t.Errorf("IsAlphabet(12345) = %t; want %t got %t", got, res, got)
	}
}

func TestIsAlphaNumeric(t *testing.T) {
	got := IsAlphaNumeric("abc")
	res := true
	if got != res {
		t.Errorf("IsAlphaNumeric(abc) = %t; want %t got %t", got, res, got)
	}

	got = IsAlphaNumeric("123")
	res = true
	if got != res {
		t.Errorf("IsAlphaNumeric(abc) = %t; want %t got %t", got, res, got)
	}

	got = IsAlphaNumeric("abc123")
	res = true
	if got != res {
		t.Errorf("IsAlphaNumeric(abc123) = %t; want %t got %t", got, res, got)
	}

	got = IsAlphaNumeric("abc123!@#")
	res = false
	if got != res {
		t.Errorf("IsEmpty(abc123!@#) = %t; want %t got %t", got, res, got)
	}

	got = IsAlphaNumeric("!@#")
	res = false
	if got != res {
		t.Errorf("IsEmpty(!@#) = %t; want %t got %t", got, res, got)
	}
}

func TestIsMobileNumber(t *testing.T) {
	got := IsMobileNumber("98461564")
	res := true
	if got != res {
		t.Errorf("IsMobileNumber(98461564) = %t; want %t got %t", got, res, got)
	}

	got = IsMobileNumber("86995154")
	res = true
	if got != res {
		t.Errorf("IsMobileNumber(86995154) = %t; want %t got %t", got, res, got)
	}

	got = IsMobileNumber("70472710")
	res = false
	if got != res {
		t.Errorf("IsMobileNumber(70472710) = %t; want %t got %t", got, res, got)
	}

	got = IsMobileNumber("8795509")
	res = false
	if got != res {
		t.Errorf("IsMobileNumber(8795509) = %t; want %t got %t", got, res, got)
	}
}
