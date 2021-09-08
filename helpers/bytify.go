package helpers

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func sizes() []string {
	return []string{"b", "K", "M", "G", "T", "P", "E", "Z"}
}

// Bytify takes a number with an ISO suffix and returns the bytes in that number as a float.
// 'size' is of the form '5G' or '0.5P'.
func Bytify(size string) (float64, error) {
	return bytifyCalc(size, 1024)
}

// BytifyI works like bytify but works with "mebibyte" type numbers.
func BytifyI(size string) (float64, error) {
	return bytifyCalc(size, 1000)
}

// UnBytify takes a number of bytes and returns that number as a string with an appropriate ISO
// suffix.
func UnBytify(size float64) string {
	return unBytifyCalc(size, 1024)
}

// UnBytifyI works like UnBytify but works with "mebibyte" type numbers.
func UnBytifyI(size float64) string {
	return unBytifyCalc(size, 1000)
}

func bytifyCalc(size string, multiplier float64) (float64, error) {
	if size == "-" {
		return 0, nil
	}

	r := regexp.MustCompile(`^\d+$`)

	if r.MatchString(size) {
		return strconv.ParseFloat(size, 64)
	}

	r = regexp.MustCompile(`^(-?[\d\.]+)(\w)$`)
	matches := r.FindAllStringSubmatch(size, -1)

	var exponent float64

	for i, v := range sizes() {
		if strings.EqualFold(v, matches[0][2]) {
			exponent = float64(i)

			break
		}
	}

	base, err := strconv.ParseFloat(matches[0][1], 64)
	if err != nil {
		return 0, err
	}

	return (base * (math.Pow(multiplier, exponent))), nil
}

func unBytifyCalc(size float64, multiplier float64) string {
	if math.Abs(size) < multiplier {
		return fmt.Sprintf("%db", int(size))
	}

	spf := ""

	if multiplier == 1000 {
		spf = "i"
	}

	for i, suffix := range sizes() {
		divisor := math.Pow(multiplier, float64(i))

		result := size / divisor

		if math.Abs(result) < multiplier {
			return fmt.Sprintf("%.1f%s%sb", result, suffix, spf)
		}
	}

	return "UNKNOWN"
}
