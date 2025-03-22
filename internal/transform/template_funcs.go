package transform

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GetTemplateFuncs returns custom template functions for transformations
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		// Time and date functions
		"now":           time.Now,
		"formatTime":    formatTime,
		"parseTime":     parseTime,
		"formatISO8601": formatISO8601,
		"formatRFC3339": formatRFC3339,
		"unixTimestamp": unixTimestamp,

		// String manipulation functions
		"lowercase":    strings.ToLower,
		"uppercase":    strings.ToUpper,
		"title":        strings.Title,
		"trim":         strings.TrimSpace,
		"trimPrefix":   strings.TrimPrefix,
		"trimSuffix":   strings.TrimSuffix,
		"replace":      strings.ReplaceAll,
		"split":        strings.Split,
		"join":         strings.Join,
		"hasPrefix":    strings.HasPrefix,
		"hasSuffix":    strings.HasSuffix,
		"contains":     strings.Contains,
		"substring":    substring,
		"regexReplace": regexReplace,
		"regexMatch":   regexMatch,
		"camelCase":    camelCase,
		"snakeCase":    snakeCase,
		"kebabCase":    kebabCase,
		"padLeft":      padLeft,
		"padRight":     padRight,

		// Encoding/decoding functions
		"base64Encode": base64Encode,
		"base64Decode": base64Decode,
		"jsonEncode":   jsonEncode,
		"jsonDecode":   jsonDecode,
		"urlEncode":    urlEncode,
		"urlDecode":    urlDecode,
		"queryEncode":  queryEncode,
		"queryDecode":  queryDecode,

		// Type conversion functions
		"toString": toString,
		"toInt":    toInt,
		"toFloat":  toFloat,
		"toBool":   toBool,
		"toJSON":   toJSON,

		// Math functions
		"add":      add,
		"subtract": subtract,
		"multiply": multiply,
		"divide":   divide,
		"modulo":   modulo,
		"max":      max,
		"min":      min,
		"round":    round,
		"floor":    floor,
		"ceil":     ceil,

		// Array/slice functions
		"first":       first,
		"last":        last,
		"rest":        rest,
		"length":      length,
		"containsVal": contains, // Renamed to avoid duplicate key
		"map":         mapFunc,
		"filter":      filterFunc,

		// Conditional functions
		"default":  defaultValue,
		"ternary":  ternary,
		"coalesce": coalesce,
		"eq":       eq,
		"ne":       ne,
		"lt":       lt,
		"le":       le,
		"gt":       gt,
		"ge":       ge,
		"and":      and,
		"or":       or,
		"not":      not,
	}
}

// Time and date functions

// formatTime formats a time according to the layout
func formatTime(layout string, t time.Time) string {
	return t.Format(layout)
}

// parseTime parses a time string according to the layout
func parseTime(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// formatISO8601 formats time in ISO8601 format
func formatISO8601(t time.Time) string {
	return t.Format("2006-01-02T15:04:05-07:00")
}

// formatRFC3339 formats time in RFC3339 format
func formatRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

// unixTimestamp returns the unix timestamp for a time
func unixTimestamp(t time.Time) int64 {
	return t.Unix()
}

// String manipulation functions

// substring returns a substring
func substring(s string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	if start > end {
		return ""
	}
	return s[start:end]
}

// regexReplace replaces text using regex
func regexReplace(pattern, replacement, text string) (string, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	return regex.ReplaceAllString(text, replacement), nil
}

// regexMatch checks if text matches a pattern
func regexMatch(pattern, text string) (bool, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	return regex.MatchString(text), nil
}

// camelCase converts a string to camelCase
func camelCase(s string) string {
	// First convert to lowercase and split by non-alphanumeric characters
	words := regexp.MustCompile("[^a-zA-Z0-9]+").Split(strings.ToLower(s), -1)

	if len(words) == 0 {
		return ""
	}

	// Start with the first word in lowercase
	result := words[0]

	// Convert remaining words to title case
	for i := 1; i < len(words); i++ {
		if words[i] != "" {
			result += strings.Title(words[i])
		}
	}

	return result
}

// snakeCase converts a string to snake_case
func snakeCase(s string) string {
	// Convert camelCase or PascalCase to snake_case
	snake := regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(s, "${1}_${2}")

	// Replace non-alphanumeric with underscore
	snake = regexp.MustCompile("[^a-zA-Z0-9]+").ReplaceAllString(snake, "_")

	// Convert to lowercase
	return strings.ToLower(strings.Trim(snake, "_"))
}

// kebabCase converts a string to kebab-case
func kebabCase(s string) string {
	// First convert to snake_case
	snake := snakeCase(s)

	// Replace underscores with hyphens
	return strings.ReplaceAll(snake, "_", "-")
}

// padLeft pads a string from the left
func padLeft(s string, padChar string, length int) string {
	if len(s) >= length {
		return s
	}
	padding := strings.Repeat(padChar, length-len(s))
	return padding + s
}

// padRight pads a string from the right
func padRight(s string, padChar string, length int) string {
	if len(s) >= length {
		return s
	}
	padding := strings.Repeat(padChar, length-len(s))
	return s + padding
}

// Encoding/decoding functions

// base64Encode encodes a string to base64
func base64Encode(v string) string {
	return base64.StdEncoding.EncodeToString([]byte(v))
}

// base64Decode decodes a base64 string
func base64Decode(v string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// jsonEncode encodes a value to JSON
func jsonEncode(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// jsonDecode decodes a JSON string
func jsonDecode(v string) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal([]byte(v), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// urlEncode encodes a string for use in a URL
func urlEncode(s string) string {
	return url.QueryEscape(s)
}

// urlDecode decodes a URL-encoded string
func urlDecode(s string) (string, error) {
	return url.QueryUnescape(s)
}

// queryEncode encodes a map to a query string
func queryEncode(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}
	return values.Encode()
}

// queryDecode decodes a query string to a map
func queryDecode(query string) (map[string]string, error) {
	values, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}

	return result, nil
}

// Type conversion functions

// toString converts a value to string
func toString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

// toInt converts a value to int
func toInt(v interface{}) (int, error) {
	switch v := v.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}

// toFloat converts a value to float64
func toFloat(v interface{}) (float64, error) {
	switch v := v.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// toBool converts a value to bool
func toBool(v interface{}) (bool, error) {
	switch v := v.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case string:
		v = strings.ToLower(strings.TrimSpace(v))
		return v == "true" || v == "yes" || v == "1" || v == "y" || v == "on", nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

// toJSON converts a value to a JSON object
func toJSON(s string) (interface{}, error) {
	var data interface{}
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Math functions

// add adds numbers
func add(a, b interface{}) (interface{}, error) {
	aFloat, err := toFloat(a)
	if err != nil {
		return nil, err
	}

	bFloat, err := toFloat(b)
	if err != nil {
		return nil, err
	}

	return aFloat + bFloat, nil
}

// subtract subtracts numbers
func subtract(a, b interface{}) (interface{}, error) {
	aFloat, err := toFloat(a)
	if err != nil {
		return nil, err
	}

	bFloat, err := toFloat(b)
	if err != nil {
		return nil, err
	}

	return aFloat - bFloat, nil
}

// multiply multiplies numbers
func multiply(a, b interface{}) (interface{}, error) {
	aFloat, err := toFloat(a)
	if err != nil {
		return nil, err
	}

	bFloat, err := toFloat(b)
	if err != nil {
		return nil, err
	}

	return aFloat * bFloat, nil
}

// divide divides numbers
func divide(a, b interface{}) (interface{}, error) {
	aFloat, err := toFloat(a)
	if err != nil {
		return nil, err
	}

	bFloat, err := toFloat(b)
	if err != nil {
		return nil, err
	}

	if bFloat == 0 {
		return nil, fmt.Errorf("division by zero")
	}

	return aFloat / bFloat, nil
}

// modulo returns the remainder of division
func modulo(a, b interface{}) (interface{}, error) {
	aInt, err := toInt(a)
	if err != nil {
		return nil, err
	}

	bInt, err := toInt(b)
	if err != nil {
		return nil, err
	}

	if bInt == 0 {
		return nil, fmt.Errorf("division by zero")
	}

	return aInt % bInt, nil
}

// max returns the maximum of values
func max(values ...interface{}) (float64, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("max requires at least one argument")
	}

	maxVal, err := toFloat(values[0])
	if err != nil {
		return 0, err
	}

	for _, v := range values[1:] {
		f, err := toFloat(v)
		if err != nil {
			return 0, err
		}
		if f > maxVal {
			maxVal = f
		}
	}

	return maxVal, nil
}

// min returns the minimum of values
func min(values ...interface{}) (float64, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("min requires at least one argument")
	}

	minVal, err := toFloat(values[0])
	if err != nil {
		return 0, err
	}

	for _, v := range values[1:] {
		f, err := toFloat(v)
		if err != nil {
			return 0, err
		}
		if f < minVal {
			minVal = f
		}
	}

	return minVal, nil
}

// round rounds a number
func round(v interface{}) (float64, error) {
	f, err := toFloat(v)
	if err != nil {
		return 0, err
	}

	return math.Round(f), nil
}

// floor returns the greatest integer value less than or equal to v
func floor(v interface{}) (float64, error) {
	f, err := toFloat(v)
	if err != nil {
		return 0, err
	}

	return math.Floor(f), nil
}

// ceil returns the least integer value greater than or equal to v
func ceil(v interface{}) (float64, error) {
	f, err := toFloat(v)
	if err != nil {
		return 0, err
	}

	return math.Ceil(f), nil
}

// Array/slice functions

// first returns the first element of a slice
func first(v interface{}) interface{} {
	switch slice := v.(type) {
	case []interface{}:
		if len(slice) == 0 {
			return nil
		}
		return slice[0]
	case []string:
		if len(slice) == 0 {
			return ""
		}
		return slice[0]
	default:
		return nil
	}
}

// last returns the last element of a slice
func last(v interface{}) interface{} {
	switch slice := v.(type) {
	case []interface{}:
		if len(slice) == 0 {
			return nil
		}
		return slice[len(slice)-1]
	case []string:
		if len(slice) == 0 {
			return ""
		}
		return slice[len(slice)-1]
	default:
		return nil
	}
}

// rest returns all but the first element of a slice
func rest(v interface{}) interface{} {
	switch slice := v.(type) {
	case []interface{}:
		if len(slice) <= 1 {
			return []interface{}{}
		}
		return slice[1:]
	case []string:
		if len(slice) <= 1 {
			return []string{}
		}
		return slice[1:]
	default:
		return []interface{}{}
	}
}

// length returns the length of a string, slice, or map
func length(v interface{}) int {
	switch v := v.(type) {
	case string:
		return len(v)
	case []interface{}:
		return len(v)
	case []string:
		return len(v)
	case map[string]interface{}:
		return len(v)
	case map[string]string:
		return len(v)
	default:
		return 0
	}
}

// contains checks if a slice contains a value
func contains(slice interface{}, value interface{}) bool {
	switch s := slice.(type) {
	case []interface{}:
		for _, v := range s {
			if v == value {
				return true
			}
		}
	case []string:
		strValue, ok := value.(string)
		if !ok {
			return false
		}
		for _, v := range s {
			if v == strValue {
				return true
			}
		}
	}
	return false
}

// mapFunc applies a function to each element of a slice
func mapFunc(slice []interface{}, fn func(interface{}) interface{}) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// filterFunc filters elements of a slice based on a predicate
func filterFunc(slice []interface{}, fn func(interface{}) bool) []interface{} {
	var result []interface{}
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Conditional functions

// defaultValue returns the default value if v is nil or empty
func defaultValue(v, d interface{}) interface{} {
	switch v := v.(type) {
	case nil:
		return d
	case string:
		if v == "" {
			return d
		}
	case int, int64, float64:
		if v == 0 {
			return d
		}
	case bool:
		if !v {
			return d
		}
	case []interface{}:
		if len(v) == 0 {
			return d
		}
	case map[string]interface{}:
		if len(v) == 0 {
			return d
		}
	case []string:
		if len(v) == 0 {
			return d
		}
	}
	return v
}

// ternary implements a ternary operator
func ternary(cond bool, trueVal, falseVal interface{}) interface{} {
	if cond {
		return trueVal
	}
	return falseVal
}

// coalesce returns the first non-empty value
func coalesce(values ...interface{}) interface{} {
	for _, v := range values {
		switch v := v.(type) {
		case nil:
			continue
		case string:
			if v != "" {
				return v
			}
		case int, int64, float64:
			if v != 0 {
				return v
			}
		case bool:
			if v {
				return v
			}
		case []interface{}:
			if len(v) > 0 {
				return v
			}
		case map[string]interface{}:
			if len(v) > 0 {
				return v
			}
		default:
			return v
		}
	}
	return nil
}

// eq checks if a equals b
func eq(a, b interface{}) bool {
	return a == b
}

// ne checks if a does not equal b
func ne(a, b interface{}) bool {
	return a != b
}

// lt checks if a is less than b
func lt(a, b interface{}) (bool, error) {
	aFloat, err := toFloat(a)
	if err != nil {
		return false, err
	}

	bFloat, err := toFloat(b)
	if err != nil {
		return false, err
	}

	return aFloat < bFloat, nil
}

// le checks if a is less than or equal to b
func le(a, b interface{}) (bool, error) {
	aFloat, err := toFloat(a)
	if err != nil {
		return false, err
	}

	bFloat, err := toFloat(b)
	if err != nil {
		return false, err
	}

	return aFloat <= bFloat, nil
}

// gt checks if a is greater than b
func gt(a, b interface{}) (bool, error) {
	aFloat, err := toFloat(a)
	if err != nil {
		return false, err
	}

	bFloat, err := toFloat(b)
	if err != nil {
		return false, err
	}

	return aFloat > bFloat, nil
}

// ge checks if a is greater than or equal to b
func ge(a, b interface{}) (bool, error) {
	aFloat, err := toFloat(a)
	if err != nil {
		return false, err
	}

	bFloat, err := toFloat(b)
	if err != nil {
		return false, err
	}

	return aFloat >= bFloat, nil
}

// and returns the logical AND of values
func and(values ...bool) bool {
	for _, v := range values {
		if !v {
			return false
		}
	}
	return true
}

// or returns the logical OR of values
func or(values ...bool) bool {
	for _, v := range values {
		if v {
			return true
		}
	}
	return false
}

// not returns the logical NOT of v
func not(v bool) bool {
	return !v
}
