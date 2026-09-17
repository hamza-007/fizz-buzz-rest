package rest

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"fizz-buzz-rest/internal/fizzbuzz"
	resp "fizz-buzz-rest/utils/resp"
)

type fizzBuzzResponse struct {
	Count  int      `json:"count"`
	Result []string `json:"result"`
}

// TODO: pure endpoint, so an ETag would work once the stats side effect moves elsewhere.
func (a *API) FizzBuzz(w http.ResponseWriter, r *http.Request) {
	req, err := parseFizzBuzzRequest(r.URL.Query(), a.maxLimit)
	if err != nil {
		resp.SendError(r.Context(), w, err)
		return
	}

	result := req.Generate()

	if r.Context().Err() != nil {
		return
	}

	a.stats.Add(req)

	resp.SendStruct(r.Context(), w, http.StatusOK, fizzBuzzResponse{
		Count:  len(result),
		Result: result,
	})
}

func parseFizzBuzzRequest(q url.Values, maxLimit int) (fizzbuzz.Request, error) {
	int1, err := queryInt(q, "int1")
	if err != nil {
		return fizzbuzz.Request{}, err
	}
	int2, err := queryInt(q, "int2")
	if err != nil {
		return fizzbuzz.Request{}, err
	}
	limit, err := queryInt(q, "limit")
	if err != nil {
		return fizzbuzz.Request{}, err
	}
	str1, err := queryString(q, "str1")
	if err != nil {
		return fizzbuzz.Request{}, err
	}
	str2, err := queryString(q, "str2")
	if err != nil {
		return fizzbuzz.Request{}, err
	}

	req := fizzbuzz.Request{Int1: int1, Int2: int2, Limit: limit, Str1: str1, Str2: str2}
	if err := req.Validate(maxLimit); err != nil {
		var fieldErr *fizzbuzz.FieldError
		if errors.As(err, &fieldErr) {
			return fizzbuzz.Request{}, resp.InvalidParameter(fieldErr.Field, fieldErr.Message)
		}
		return fizzbuzz.Request{}, err
	}
	return req, nil
}

func queryInt(q url.Values, name string) (int, error) {
	raw, err := queryString(q, name)
	if err != nil {
		return 0, err
	}
	value, convErr := strconv.Atoi(raw)
	if convErr != nil {
		return 0, resp.InvalidParameter(name, "must be a valid integer")
	}
	return value, nil
}

func queryString(q url.Values, name string) (string, error) {
	if !q.Has(name) {
		return "", resp.MissingParameter(name)
	}
	return q.Get(name), nil
}
