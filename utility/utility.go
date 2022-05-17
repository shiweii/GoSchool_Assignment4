// Package utility implements various functionalities shared between various packages
package utility

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shiweii/logger"

	"github.com/joho/godotenv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// CurrFuncName return the function name which this function was called
// used mainly in logging to determine which function the log was called.
func CurrFuncName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function
}

// GetEnvVar read all vars declared in .env.
func GetEnvVar(v string) string {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal.Fatal("Error loading .env file")
	}
	return os.Getenv(v)
}

// VerifyCheckSum verify that file was not tempered with by checking
// against the checksum of the file.
func VerifyCheckSum() {
	for {
		logChecksum, err := ioutil.ReadFile(GetEnvVar("CHECKSUM_FILE"))
		if err != nil {
			logger.Error.Println(err)
		}
		// convert content to a 'string'
		str := string(logChecksum)
		// Compute our current log's SHA256 hash
		hash, err := computeSHA512(GetEnvVar("CHECKSUM_FILE_TO_VERIFY"))
		if err != nil {
			logger.Error.Println(err)
		} else {
			// Compare our calculated hash with our stored hash
			if str == hash {
				// Ok the checksums match.
				logger.Info.Println("File integrity OK.")
			} else {
				// The file integrity has been compromised...
				logger.Warning.Println("File Tampering detected.")
			}
		}
		min, _ := time.ParseDuration(GetEnvVar("CHECKSUM_TIMER"))
		time.Sleep(min)
	}
}

// GetEnvVar computes the SHA512 checksum of a given file.
func computeSHA512(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logger.Error.Println(err)
		}
	}(f)

	harsher := sha512.New()
	if _, err := io.Copy(harsher, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(harsher.Sum(nil)), nil
}

// LevenshteinDistance computes and returns
// the number of changes between two strings.
func LevenshteinDistance(s, t string) int {
	// Change string to lower case for accurate comparison
	s = strings.ToLower(s)
	t = strings.ToLower(t)
	// Create LD Matrix
	d := make([][]int, len(t)+1)
	for i := range d {
		d[i] = make([]int, len(s)+1)
	}
	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	// Loop LD Matrix
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[j][i] = d[j-1][i-1]
			} else {
				// Check for Min
				min := d[j-1][i-1]
				if d[j][i-1] < min {
					min = d[j][i-1]
				}
				if d[j-1][i] < min {
					min = d[j-1][i]
				}
				d[j][i] = min + 1
			}
		}
	}
	return d[len(t)][len(s)]
}

// GenerateID generates a random number using math/rand
// do not use if security is needed.
func GenerateID() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(10000000000)
}

// AddOne return plus 1 to input integer.
func AddOne(x int) int {
	return x + 1
}

// FirstCharToUpper changes string to tile case.
func FirstCharToUpper(x string) string {
	return cases.Title(language.Und, cases.NoLower).String(x)
}

// FormatDate parse and format date to YYYY-MM-DD format.
func FormatDate(x string) string {
	td, err := time.Parse("2006-01-02", x)
	if err != nil {
		logger.Error.Println(err)
	} else {
		return td.Format("02-Jan-2006")
	}
	return ""
}

// GetDay parse and returns the day of a given date.
func GetDay(x string) string {
	td, err := time.Parse("2006-01-02", x)
	if err != nil {
		fmt.Println(err)
	} else {
		return td.Weekday().String()
	}
	return ""
}
