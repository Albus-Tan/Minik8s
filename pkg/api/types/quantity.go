package types

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Quantity such as "1500m", "1200M" or "1"
type Quantity string

// ParseQuantity parse Quantity and return uint64 whose unit is m/M
func ParseQuantity(name ResourceName, q Quantity) (uint64, error) {
	if len(q) == 0 {
		return 0, errors.New("quantity empty string")
	}
	if q == "0" {
		return 0, nil
	}

	positive, value, _, _, suf, err := parseQuantityString(string(q))
	if err != nil {
		return 0, err
	}
	if !positive {
		return 0, errors.New("quantity not positive")
	}

	switch name {
	case ResourceCPU:
		if suf == "m" || suf == "M" {
			return strconv.ParseUint(value, 10, 64)
		} else if suf == "" {
			return strconv.ParseUint(value+"000", 10, 64)
		}
	case ResourceMemory:
		if suf == "m" || suf == "M" {
			return strconv.ParseUint(value, 10, 64)
		} else if suf == "" {
			return strconv.ParseUint(value, 10, 64)
		}
	default:
		return 0, errors.New(fmt.Sprintf("resource type %v unsupported", name))
	}
	return 0, errors.New(fmt.Sprintf("suf unit %v unsupported", suf))
}

// parseQuantityString is a fast scanner for quantity values.
func parseQuantityString(str string) (positive bool, value, num, denom, suffix string, err error) {
	positive = true
	pos := 0
	end := len(str)

	// handle leading sign
	if pos < end {
		switch str[0] {
		case '-':
			positive = false
			pos++
		case '+':
			pos++
		}
	}

	// strip leading zeros
Zeroes:
	for i := pos; ; i++ {
		if i >= end {
			num = "0"
			value = num
			return
		}
		switch str[i] {
		case '0':
			pos++
		default:
			break Zeroes
		}
	}

	// extract the numerator
Num:
	for i := pos; ; i++ {
		if i >= end {
			num = str[pos:end]
			value = str[0:end]
			return
		}
		switch str[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			num = str[pos:i]
			pos = i
			break Num
		}
	}

	// if we stripped all numerator positions, always return 0
	if len(num) == 0 {
		num = "0"
	}

	// handle a denominator
	if pos < end && str[pos] == '.' {
		pos++
	Denom:
		for i := pos; ; i++ {
			if i >= end {
				denom = str[pos:end]
				value = str[0:end]
				return
			}
			switch str[i] {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			default:
				denom = str[pos:i]
				pos = i
				break Denom
			}
		}
	}
	value = str[0:pos]

	// grab the elements of the suffix
	suffixStart := pos
	for i := pos; ; i++ {
		if i >= end {
			suffix = str[suffixStart:end]
			return
		}
		if !strings.ContainsAny(str[i:i+1], "eEinumkKMGTP") {
			pos = i
			break
		}
	}
	if pos < end {
		switch str[pos] {
		case '-', '+':
			pos++
		}
	}
Suffix:
	for i := pos; ; i++ {
		if i >= end {
			suffix = str[suffixStart:end]
			return
		}
		switch str[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			break Suffix
		}
	}
	// we encountered a non decimal in the Suffix loop, but the last character
	// was not a valid exponent
	err = errors.New("parseQuantityString failed")
	return
}
