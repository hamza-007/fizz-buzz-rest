package fizzbuzz

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

const maxStringLen = 100

const maxPrealloc = 1 << 16

type Request struct {
	Int1  int    `json:"int1"`
	Int2  int    `json:"int2"`
	Limit int    `json:"limit"`
	Str1  string `json:"str1"`
	Str2  string `json:"str2"`
}

type FieldError struct {
	Field   string
	Message string
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (r Request) Validate(maxLimit int) error {
	if r.Int1 <= 0 {
		return &FieldError{Field: "int1", Message: "must be a positive integer"}
	}
	if r.Int2 <= 0 {
		return &FieldError{Field: "int2", Message: "must be a positive integer"}
	}
	if r.Limit < 0 {
		return &FieldError{Field: "limit", Message: "must be zero or a positive integer"}
	}
	if maxLimit > 0 && r.Limit > maxLimit {
		return &FieldError{Field: "limit", Message: fmt.Sprintf("must not exceed %d", maxLimit)}
	}
	if err := validateReplacement("str1", r.Str1); err != nil {
		return err
	}
	return validateReplacement("str2", r.Str2)
}

func validateReplacement(field, value string) error {
	if value == "" {
		return &FieldError{Field: field, Message: "must not be empty"}
	}
	if utf8.RuneCountInString(value) > maxStringLen {
		return &FieldError{Field: field, Message: fmt.Sprintf("must not exceed %d characters", maxStringLen)}
	}
	return nil
}

func (r Request) Generate() []string {
	out := make([]string, 0, min(r.Limit, maxPrealloc))
	both := r.Str1 + r.Str2

	for i := 1; i <= r.Limit; i++ {
		switch {
		case i%r.Int1 == 0 && i%r.Int2 == 0:
			out = append(out, both)
		case i%r.Int1 == 0:
			out = append(out, r.Str1)
		case i%r.Int2 == 0:
			out = append(out, r.Str2)
		default:
			out = append(out, strconv.Itoa(i))
		}
	}
	return out
}
