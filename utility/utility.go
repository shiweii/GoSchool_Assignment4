package utility

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/shiweii/logger"

	"github.com/joho/godotenv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetEnvVar(v string) string {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal.Fatal("Error loading .env file")
	}
	return os.Getenv(v)
}

func VerifyCheckSum() {
	for {
		logChecksum, err := ioutil.ReadFile(GetEnvVar("CHECKSUM_FILE"))
		if err != nil {
			logger.Error.Println(err)
		}
		// convert content to a 'string'
		str := string(logChecksum)
		// Compute our current log's SHA256 hash
		hash, err := computeSHA256(GetEnvVar("CHECKSUM_FILE_TO_VERIFY"))
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
		//env := GetEnvVar("CHECKSUM_TIMER")
		//timer, _ := strconv.Atoi(env)
		time.Sleep(10 * time.Minute)
	}
}

func computeSHA256(file string) (string, error) {
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

	harsher := sha256.New()
	if _, err := io.Copy(harsher, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(harsher.Sum(nil)), nil
}

// ReadInputAsInt Read input and parse as int, false if user entered non integer

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

func GenerateID() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(10000000000)
}

func AddOne(x int) int {
	return x + 1
}

func FirstCharToUpper(x string) string {
	return cases.Title(language.Und, cases.NoLower).String(x)
}

func FormatDate(x string) string {
	td, err := time.Parse("2006-01-02", x)
	if err != nil {
		fmt.Println(err)
	} else {
		return td.Format("02-Jan-2006")
	}
	return ""
}

func GetDay(x string) string {
	td, err := time.Parse("2006-01-02", x)
	if err != nil {
		fmt.Println(err)
	} else {
		return td.Weekday().String()
	}
	return ""
}
