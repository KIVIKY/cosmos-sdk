package schema

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"
)

// Kind represents the basic type of a field in an object.
// Each kind defines the following encodings:
//
// - Go Encoding: the golang type which should be accepted by listeners and
// generated by decoders when providing entity updates.
// - JSON Encoding: the JSON encoding which should be used when encoding the field to JSON.
// - Key Wire Encoding: the encoding which should be used when encoding the field
// as a key in binary messages. Some encodings specify a terminal and non-terminal form
// depending on whether or not the field is the last field in the key.
// - Value Wire Encoding: the encoding which should be used when encoding the field
// as a value in binary messages.
//
// When there is some non-determinism in an encoding, kinds should specify what
// values they accept and also what is the canonical, deterministic encoding which
// should be preferably emitted by serializers.
//
// Wire encodings were chosen based on what is likely to be the most convenient default binary encoding
// in key-value stores so that implementations that chose the wire encoding natively have a trivial
// decoder implementation.
type Kind int

const (
	// InvalidKind indicates that an invalid type.
	InvalidKind Kind = iota

	// StringKind is a string type.
	// Go Encoding: UTF-8 string with no null characters.
	// JSON Encoding: string
	// Key Wire Encoding:
	//   UTF-8 string with no null characters suffixed with a null character (non-terminal)
	//   UTF-8 string with no null characters (terminal)
	// Value Wire Encoding: UTF-8 string with no null characters suffixed with a null character
	StringKind

	// BytesKind is a bytes type.
	// Go Encoding: []byte
	// JSON Encoding: base64 encoded string, canonical values should be encoded with standard encoding and padding.
	// Either standard or URL encoding with or without padding should be accepted.
	BytesKind

	// Int8Kind represents an 8-bit signed integer.
	// Go Encoding: int8
	// JSON Encoding: number
	Int8Kind

	// Uint8Kind represents an 8-bit unsigned integer.
	// Go Encoding: uint8
	// JSON Encoding: number
	Uint8Kind

	// Int16Kind represents a 16-bit signed integer.
	// Go Encoding: int16
	// JSON Encoding: number
	Int16Kind

	// Uint16Kind represents a 16-bit unsigned integer.
	// Go Encoding: uint16
	// JSON Encoding: number
	Uint16Kind

	// Int32Kind represents a 32-bit signed integer.
	// Go Encoding: int32
	// JSON Encoding: number
	Int32Kind

	// Uint32Kind represents a 32-bit unsigned integer.
	// Go Encoding: uint32
	// JSON Encoding: number
	Uint32Kind

	// Int64Kind represents a 64-bit signed integer.
	// Go Encoding: int64
	// JSON Encoding: base10 integer string which matches the IntegerFormat regex
	// The canonical encoding should include no leading zeros.
	Int64Kind

	// Uint64Kind represents a 64-bit unsigned integer.
	// Go Encoding: uint64
	// JSON Encoding: base10 integer string which matches the IntegerFormat regex
	// Canonically encoded values should include no leading zeros.
	Uint64Kind

	// IntegerStringKind represents an arbitrary precision integer number.
	// Go Encoding: string which matches the IntegerFormat regex
	// JSON Encoding: base10 integer string
	// Canonically encoded values should include no leading zeros.
	IntegerStringKind

	// DecimalStringKind represents an arbitrary precision decimal or integer number.
	// Go Encoding: string which matches the DecimalFormat regex
	// JSON Encoding: base10 decimal string
	// Canonically encoded values should include no leading zeros or trailing zeros,
	// and exponential notation with a lowercase 'e' should be used for any numbers
	// with an absolute value less than or equal to 1e-6 or greater than or equal to 1e6.
	DecimalStringKind

	// BoolKind represents a boolean true or false value.
	// Go Encoding: bool
	// JSON Encoding: boolean
	BoolKind

	// TimeKind represents a nanosecond precision UNIX time value (with zero representing January 1, 1970 UTC).
	// Its valid range is +/- 2^63 (the range of a 64-bit signed integer).
	// Go Encoding: time.Time
	// JSON Encoding: Any value IS0 8601 time stamp should be accepted.
	// Canonical values should be encoded with UTC time zone Z, nanoseconds should
	// be encoded with no trailing zeros, and T time values should always be present
	// even at 00:00:00.
	TimeKind

	// DurationKind represents the elapsed time between two nanosecond precision time values.
	// Its valid range is +/- 2^63 (the range of a 64-bit signed integer).
	// Go Encoding: time.Duration
	// JSON Encoding: the number of seconds as a decimal string with no trailing zeros followed by
	// a lowercase 's' character to represent seconds.
	DurationKind

	// Float32Kind represents an IEEE-754 32-bit floating point number.
	// Go Encoding: float32
	// JSON Encoding: number
	Float32Kind

	// Float64Kind represents an IEEE-754 64-bit floating point number.
	// Go Encoding: float64
	// JSON Encoding: number
	Float64Kind

	// AddressKind represents an account address which is represented by a variable length array of bytes.
	// Addresses usually have a human-readable rendering, such as bech32, and tooling should provide
	// a way for apps to define a string encoder for friendly user-facing display.
	// Go Encoding: []byte
	// JSON Encoding: addresses should be encoded as strings using the human-readable address renderer
	// provided to the JSON encoder.
	AddressKind

	// EnumKind represents a value of an enum type.
	// Fields of this type are expected to set the EnumType field in the field definition to the enum
	// definition.
	// Go Encoding: string
	// JSON Encoding: string
	EnumKind

	// JSONKind represents arbitrary JSON data.
	// Go Encoding: json.RawMessage
	// JSON Encoding: any valid JSON value
	JSONKind
)

// MAX_VALID_KIND is the maximum valid kind value.
const MAX_VALID_KIND = JSONKind

const (
	// IntegerFormat is a regex that describes the format integer number strings must match. It specifies
	// that integers may have at most 100 digits.
	IntegerFormat = `^-?[0-9]{1,100}$`

	// DecimalFormat is a regex that describes the format decimal number strings must match. It specifies
	// that decimals may have at most 50 digits before and after the decimal point and may have an optional
	// exponent of up to 2 digits. These restrictions ensure that the decimal can be accurately represented
	// by a wide variety of implementations.
	DecimalFormat = `^-?[0-9]{1,50}(\.[0-9]{1,50})?([eE][-+]?[0-9]{1,2})?$`
)

// Validate returns an errContains if the kind is invalid.
func (t Kind) Validate() error {
	if t <= InvalidKind {
		return fmt.Errorf("unknown type: %d", t)
	}
	if t > JSONKind {
		return fmt.Errorf("invalid type: %d", t)
	}
	return nil
}

// String returns a string representation of the kind.
func (t Kind) String() string {
	switch t {
	case StringKind:
		return "string"
	case BytesKind:
		return "bytes"
	case Int8Kind:
		return "int8"
	case Uint8Kind:
		return "uint8"
	case Int16Kind:
		return "int16"
	case Uint16Kind:
		return "uint16"
	case Int32Kind:
		return "int32"
	case Uint32Kind:
		return "uint32"
	case Int64Kind:
		return "int64"
	case Uint64Kind:
		return "uint64"
	case DecimalStringKind:
		return "decimal"
	case IntegerStringKind:
		return "integer"
	case BoolKind:
		return "bool"
	case TimeKind:
		return "time"
	case DurationKind:
		return "duration"
	case Float32Kind:
		return "float32"
	case Float64Kind:
		return "float64"
	case AddressKind:
		return "address"
	case EnumKind:
		return "enum"
	case JSONKind:
		return "json"
	default:
		return fmt.Sprintf("invalid(%d)", t)
	}
}

// ValidateValueType returns an errContains if the value does not conform to the expected go type.
// Some fields may accept nil values, however, this method does not have any notion of
// nullability. This method only validates that the go type of the value is correct for the kind
// and does not validate string or json formats. Kind.ValidateValue does a more thorough validation
// of number and json string formatting.
func (t Kind) ValidateValueType(value interface{}) error {
	switch t {
	case StringKind:
		_, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case BytesKind:
		_, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("expected []byte, got %T", value)
		}
	case Int8Kind:
		_, ok := value.(int8)
		if !ok {
			return fmt.Errorf("expected int8, got %T", value)
		}
	case Uint8Kind:
		_, ok := value.(uint8)
		if !ok {
			return fmt.Errorf("expected uint8, got %T", value)
		}
	case Int16Kind:
		_, ok := value.(int16)
		if !ok {
			return fmt.Errorf("expected int16, got %T", value)
		}
	case Uint16Kind:
		_, ok := value.(uint16)
		if !ok {
			return fmt.Errorf("expected uint16, got %T", value)
		}
	case Int32Kind:
		_, ok := value.(int32)
		if !ok {
			return fmt.Errorf("expected int32, got %T", value)
		}
	case Uint32Kind:
		_, ok := value.(uint32)
		if !ok {
			return fmt.Errorf("expected uint32, got %T", value)
		}
	case Int64Kind:
		_, ok := value.(int64)
		if !ok {
			return fmt.Errorf("expected int64, got %T", value)
		}
	case Uint64Kind:
		_, ok := value.(uint64)
		if !ok {
			return fmt.Errorf("expected uint64, got %T", value)
		}
	case IntegerStringKind:
		_, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}

	case DecimalStringKind:
		_, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case BoolKind:
		_, ok := value.(bool)
		if !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	case TimeKind:
		_, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("expected time.Time, got %T", value)
		}
	case DurationKind:
		_, ok := value.(time.Duration)
		if !ok {
			return fmt.Errorf("expected time.Duration, got %T", value)
		}
	case Float32Kind:
		_, ok := value.(float32)
		if !ok {
			return fmt.Errorf("expected float32, got %T", value)
		}
	case Float64Kind:
		_, ok := value.(float64)
		if !ok {
			return fmt.Errorf("expected float64, got %T", value)
		}
	case AddressKind:
		_, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("expected []byte, got %T", value)
		}
	case EnumKind:
		_, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case JSONKind:
		_, ok := value.(json.RawMessage)
		if !ok {
			return fmt.Errorf("expected json.RawMessage, got %T", value)
		}
	default:
		return fmt.Errorf("invalid type: %d", t)
	}
	return nil
}

// ValidateValue returns an errContains if the value does not conform to the expected go type and format.
// It is more thorough, but slower, than Kind.ValidateValueType and validates that Integer, Decimal and JSON
// values are formatted correctly. It cannot validate enum values because Kind's do not have enum schemas.
func (t Kind) ValidateValue(value interface{}) error {
	err := t.ValidateValueType(value)
	if err != nil {
		return err
	}

	switch t {
	case StringKind:
		str := value.(string)
		if !utf8.ValidString(str) {
			return fmt.Errorf("expected valid utf-8 string, got %s", value)
		}

		// check for null characters
		for _, r := range str {
			if r == 0 {
				return fmt.Errorf("expected string without null characters, got %s", value)
			}
		}
	case IntegerStringKind:
		if !integerRegex.Match([]byte(value.(string))) {
			return fmt.Errorf("expected base10 integer, got %s", value)
		}
	case DecimalStringKind:
		if !decimalRegex.Match([]byte(value.(string))) {
			return fmt.Errorf("expected decimal number, got %s", value)
		}
	case JSONKind:
		if !json.Valid(value.(json.RawMessage)) {
			return fmt.Errorf("expected valid JSON, got %s", value)
		}
	default:
		return nil
	}
	return nil
}

var (
	integerRegex = regexp.MustCompile(IntegerFormat)
	decimalRegex = regexp.MustCompile(DecimalFormat)
)

// KindForGoValue finds the simplest kind that can represent the given go value. It will not, however,
// return kinds such as IntegerStringKind, DecimalStringKind, AddressKind, or EnumKind which all can be
// represented as strings.
func KindForGoValue(value interface{}) Kind {
	switch value.(type) {
	case string:
		return StringKind
	case []byte:
		return BytesKind
	case int8:
		return Int8Kind
	case uint8:
		return Uint8Kind
	case int16:
		return Int16Kind
	case uint16:
		return Uint16Kind
	case int32:
		return Int32Kind
	case uint32:
		return Uint32Kind
	case int64:
		return Int64Kind
	case uint64:
		return Uint64Kind
	case float32:
		return Float32Kind
	case float64:
		return Float64Kind
	case bool:
		return BoolKind
	case time.Time:
		return TimeKind
	case time.Duration:
		return DurationKind
	case json.RawMessage:
		return JSONKind
	default:
		return InvalidKind
	}
}
