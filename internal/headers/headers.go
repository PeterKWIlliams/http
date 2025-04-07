package headers

import (
	"errors"
	"fmt"
	"strings"
)

var allowedCharSet map[rune]struct{}

func init() {
	allowedCharSet = make(map[rune]struct{})
	for r := 'A'; r <= 'Z'; r++ {
		allowedCharSet[r] = struct{}{}
	}

	for r := 'a'; r <= 'z'; r++ {
		allowedCharSet[r] = struct{}{}
	}

	for r := '0'; r <= '9'; r++ {
		allowedCharSet[r] = struct{}{}
	}

	specialChars := `!#$%&'*+-.^_` + "`|~"
	for _, r := range specialChars {
		allowedCharSet[r] = struct{}{}
	}
}

type Headers map[string]string

func NewHeaders() Headers { return Headers{} }

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	headerStr := string(data)
	if !strings.Contains(headerStr, "\r\n") {
		return 0, false, nil
	}

	if strings.HasPrefix(headerStr, "\r\n") {
		return 2, true, nil
	}
	headerStr = strings.Split(headerStr, "\r\n")[0]
	consumedBytes := len(headerStr) + 2
	headersParts := strings.SplitN(headerStr, ":", 2)

	if len(headersParts) < 2 {
		return 0, false, fmt.Errorf("Invalid header: missing ':' separator")
	}

	fieldName := strings.TrimLeft(headersParts[0], " ")
	fieldValue := strings.TrimSpace(headersParts[1])

	if err = validFieldName(fieldName); err != nil {
		return 0, false, fmt.Errorf("Invalid field name: %s", err)
	}
	fieldName = strings.ToLower(fieldName)

	if fv, exists := h[fieldName]; exists {
		h[fieldName] = fmt.Sprintf("%s, %s", fv, fieldValue)
	} else {
		h[fieldName] = fieldValue
	}

	return consumedBytes, false, nil
}

func (h Headers) Get(fieldName string) (string, error) {
	canonicalFieldName := strings.ToLower(fieldName)
	if val, exists := h[canonicalFieldName]; exists {
		return val, nil
	}
	return "", errors.New("header field does not exist")
}

func validFieldName(fieldName string) error {
	for _, char := range fieldName {
		if _, exists := allowedCharSet[char]; !exists {
			return fmt.Errorf("Header field name '%s' contains invalid character '%s'", fieldName, string(char))
		}
	}
	if fieldName == "" {
		return fmt.Errorf("Header field name cannot be empty")
	}
	if strings.Contains(fieldName, " ") {
		return fmt.Errorf("Whitespace between field name and field value:%s", fieldName)
	}
	return nil
}
