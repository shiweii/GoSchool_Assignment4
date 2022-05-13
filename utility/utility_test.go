package utility

import (
	"testing"
)

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
