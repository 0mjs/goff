package eval

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// EvalOperator evaluates an attribute condition.
// Returns (match, error). If error is non-nil, the condition should be skipped.
func EvalOperator(attrValue interface{}, op string, opValue interface{}, compiledRegex *regexp.Regexp) (bool, error) {
	switch op {
	case "eq":
		return eq(attrValue, opValue)
	case "neq":
		match, err := eq(attrValue, opValue)
		return !match, err
	case "gt":
		return compare(attrValue, opValue, 1)
	case "gte":
		return compare(attrValue, opValue, 0, 1)
	case "lt":
		return compare(attrValue, opValue, -1)
	case "lte":
		return compare(attrValue, opValue, -1, 0)
	case "in":
		return in(attrValue, opValue)
	case "contains":
		return contains(attrValue, opValue)
	case "matches":
		return matches(attrValue, compiledRegex)
	default:
		return false, fmt.Errorf("unknown operator: %s", op)
	}
}

func eq(attrValue, opValue interface{}) (bool, error) {
	// Direct comparison first
	if attrValue == opValue {
		return true, nil
	}

	// Try numeric comparison
	attrNum, attrOk := toFloat64(attrValue)
	opNum, opOk := toFloat64(opValue)
	if attrOk && opOk {
		return attrNum == opNum, nil
	}

	// Try string comparison
	attrStr := fmt.Sprintf("%v", attrValue)
	opStr := fmt.Sprintf("%v", opValue)
	return attrStr == opStr, nil
}

func compare(attrValue, opValue interface{}, allowedResults ...int) (bool, error) {
	attrNum, attrOk := toFloat64(attrValue)
	opNum, opOk := toFloat64(opValue)

	if !attrOk || !opOk {
		return false, fmt.Errorf("comparison operators require numeric values")
	}

	result := 0
	if attrNum > opNum {
		result = 1
	} else if attrNum < opNum {
		result = -1
	}

	for _, allowed := range allowedResults {
		if result == allowed {
			return true, nil
		}
	}
	return false, nil
}

func in(attrValue, opValue interface{}) (bool, error) {
	// opValue should be a slice/array
	opSlice, ok := opValue.([]interface{})
	if !ok {
		return false, fmt.Errorf("'in' operator requires array value")
	}

	attrStr := fmt.Sprintf("%v", attrValue)
	for _, v := range opSlice {
		if fmt.Sprintf("%v", v) == attrStr {
			return true, nil
		}
	}
	return false, nil
}

func contains(attrValue, opValue interface{}) (bool, error) {
	attrStr := fmt.Sprintf("%v", attrValue)
	opStr := fmt.Sprintf("%v", opValue)
	return strings.Contains(attrStr, opStr), nil
}

func matches(attrValue interface{}, regex *regexp.Regexp) (bool, error) {
	if regex == nil {
		return false, fmt.Errorf("regex not compiled")
	}
	attrStr := fmt.Sprintf("%v", attrValue)
	return regex.MatchString(attrStr), nil
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

