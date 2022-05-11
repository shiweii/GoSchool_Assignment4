package utility

import (
	"testing"
)

func TestValidateMobileNumber(t *testing.T) {
	got := ValidateMobileNumber(98461564)
	res := true
	if got != res {
		t.Errorf("ValidateMobileNumber(98461564) = %t; want %t got %t", got, res, got)
	}

	got = ValidateMobileNumber(86995154)
	res = true
	if got != res {
		t.Errorf("ValidateMobileNumber(86995154) = %t; want %t got %t", got, res, got)
	}

	got = ValidateMobileNumber(70472710)
	res = false
	if got != res {
		t.Errorf("ValidateMobileNumber(70472710) = %t; want %t got %t", got, res, got)
	}

	got = ValidateMobileNumber(8795509)
	res = false
	if got != res {
		t.Errorf("ValidateMobileNumber(8795509) = %t; want %t got %t", got, res, got)
	}
}

func TestLevenshteinDistance(t *testing.T) {
	got := LevenshteinDistance("kitten", "sitting")
	res := 3
	if got != res {
		t.Errorf("LevenshteinDistance(kitten, sitting) = %d; want %d got %d", got, res, got)
	}

	got = LevenshteinDistance("rosettacode", "raisethysword")
	res = 8
	if got != res {
		t.Errorf("LevenshteinDistance(rosettacode, raisethysword) = %d; want %d got %d", got, res, got)
	}

	got = LevenshteinDistance("Diana Prince", "DIaNa pRincE")
	res = 0
	if got != res {
		t.Errorf("LevenshteinDistance(Bruce Wayne, bruce wayne) = %d; want %d got %d", got, res, got)
	}

	got = LevenshteinDistance("Bruce Wayne", "bruce wayne")
	res = 0
	if got != res {
		t.Errorf("LevenshteinDistance(Bruce Wayne, bruce wayne) = %d; want %d got %d", got, res, got)
	}

	got = LevenshteinDistance("Clark Kent", "Clark, Kant")
	res = 2
	if got != res {
		t.Errorf("LevenshteinDistance(Clark Kent, Clark, Kant) = %d; want %d got %d", got, res, got)
	}
}
