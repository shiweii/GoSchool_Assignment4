package utility

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetEnvVar(v string) string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return os.Getenv(v)
}

func ReadInput() string {
	inputReader := bufio.NewReader(os.Stdin)
	v, _ := inputReader.ReadString('\n')
	v = strings.TrimSpace(v)
	return v
}

// Read input and parse as int, false if user entered non integer
func ReadInputAsInt() (int, error) {
	inputReader := bufio.NewReader(os.Stdin)
	input, err := inputReader.ReadString('\n')
	if err != nil {
		return 0, err
	}

	input = strings.TrimSpace(input)
	// Check that user input valid selection
	if value, err := strconv.Atoi(input); err == nil {
		return value, nil
	} else {
		return 0, errors.New("please enter a valid selection")
	}
}

func ValidateMobileNumber(mobileNum int) bool {
	var ret bool = false
	str := strconv.Itoa(mobileNum)
	if len(str) == 8 {
		if str[0:1] == "8" || str[0:1] == "9" {
			ret = true
		}
	}
	return ret
}

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
