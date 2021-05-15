package helpers

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Bytify takes a number with an ISO suffix and returns the bytes in that number as a float.
// 'size' is of the form '5G' or '0.5P'.
func Bytify(size string) (float64, error) {
	return bytifyCalc(size, 1024)
}

// BytifyI works like bytify but works with "mebibyte" type numbers.
func BytifyI(size string) (float64, error) {
	return bytifyCalc(size, 1000)
}

func bytifyCalc(size string, multiplier float64) (float64, error) {
	sizes := [8]string{"b", "K", "M", "G", "T", "P", "E", "Z"}

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

	for i, v := range sizes {
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
