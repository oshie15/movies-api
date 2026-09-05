package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Define an error that our UnmarshalJSON() method can return if we are unable to parse
// or convert the JSON string successfully.
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	// Use the strconv.Quote() function on the string to wrap it in double quotes.
	// it needs to be surronded by double quotes in order to be valid *JSON string*.
	quotedJSONValue := strconv.Quote(jsonValue)

	// Convert the quoted string value to a byte slice and return it.
	return []byte(quotedJSONValue), nil
}

// Implement a UnmarshallJSON() method on the Runtime type so that it satisifies the
// json.Unmarshaler interface. IMPORTANT: Because UnmarshallJSON needs to modify
// reciver, we must use pointer  receiver for this to work
// correctly, otherwise, we will only be modifying a copy
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// We expect thta the incoming JSON value will be string in the format
	// "<runtime> mins" and the first thing  we need to do is remove the surronding
	// double quotes from this string. If we can not unquote it, then we return the
	// ErrInvalidRuntimeFormat error.
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Split the string to isolate the part containing the number
	parts := strings.Split(unquotedJSONValue, " ")

	// Sanity check the parts of the string to make sure it was in the expected format.
	// If it isnot, we return the ErrInvalidRuntimeFormat error again.
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// Otherwise, parse the string containing the number into an int32. Again, if this
	// fails return the ErrorInvalidRuntimeFormat error
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Convert the int32 to a Runtime type and assign this to the receiver. Note that we
	// use the * operator to dereference the receiver in order to set
	// the underlying value of the pointer.
	*r = Runtime(i)

	return nil
}
