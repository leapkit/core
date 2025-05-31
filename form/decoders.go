package form

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"
)

var (
	mu             sync.Mutex
	customDecoders = map[reflect.Type]func(string) (any, error){
		reflect.TypeOf(time.Time{}): func(value string) (any, error) {
			layouts := []string{
				time.Layout,
				time.ANSIC,
				time.UnixDate,
				time.RubyDate,
				time.RFC822,
				time.RFC822Z,
				time.RFC850,
				time.RFC1123,
				time.RFC1123Z,
				time.RFC3339,
				time.RFC3339Nano,
				time.Kitchen,
				time.Stamp,
				time.StampMilli,
				time.StampMicro,
				time.StampNano,
				time.DateTime,
				time.DateOnly,
				time.TimeOnly,
			}

			for _, l := range layouts {
				t, err := time.Parse(l, value)
				if err != nil {
					continue
				}

				return t, nil
			}

			return nil, fmt.Errorf("invalid time value: %q", value)
		},
	}
)

func RegisterCustomTypeFunc(fn func(string) (any, error), customType any) {
	mu.Lock()
	defer mu.Unlock()

	t := reflect.TypeOf(customType)
	customDecoders[t] = fn
}

var builtInDecoders = map[reflect.Kind]func(string) (any, error){
	reflect.Bool: func(value string) (any, error) {
		v, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("invalid bool value: %q", value)
		}
		return v, nil
	},
	reflect.Int: func(value string) (any, error) {
		v, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("invalid int value: %q", value)
		}
		return int(v), nil
	},
	reflect.Int8: func(value string) (any, error) {
		v, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid int8 value: %q", value)
		}
		return int8(v), nil
	},
	reflect.Int16: func(value string) (any, error) {
		v, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid int16 value: %q", value)
		}
		return int16(v), nil
	},
	reflect.Int32: func(value string) (any, error) {
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid int32 value: %q", value)
		}
		return int32(v), nil
	},
	reflect.Int64: func(value string) (any, error) {
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid int64 value: %q", value)
		}
		return v, nil
	},
	reflect.Uint: func(value string) (any, error) {
		v, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("invalid uint value: %q", value)
		}
		return uint(v), nil
	},
	reflect.Uint8: func(value string) (any, error) {
		v, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid uint8 value: %q", value)
		}
		return uint8(v), nil
	},
	reflect.Uint16: func(value string) (any, error) {
		v, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid uint16 value: %q", value)
		}
		return uint16(v), nil
	},
	reflect.Uint32: func(value string) (any, error) {
		v, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid uint32 value: %q", value)
		}
		return uint32(v), nil
	},
	reflect.Uint64: func(value string) (any, error) {
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid uint64 value: %q", value)
		}
		return v, nil
	},
	reflect.Float32: func(value string) (any, error) {
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid float32 value: %q", value)
		}
		return float32(v), nil
	},
	reflect.Float64: func(value string) (any, error) {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float64 value: %q", value)
		}
		return v, nil
	},
	reflect.String: func(value string) (any, error) {
		return value, nil
	},
}
