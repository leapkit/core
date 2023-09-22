package form

import (
	"fmt"
	"time"

	"github.com/go-playground/form/v4"
)

// use a single instance of Decoder, it caches struct info
var (
	defaultDecoder = newDecoder()

	Decode = defaultDecoder.decode

	// RegisterTypeDecoder allows to define how certain types will
	// be decoded from string values.
	RegisterTypeDecoder = defaultDecoder.RegisterCustomTypeFunc
)

func newDecoder() *decoder {
	dec := &decoder{
		form.NewDecoder(),
	}

	dec.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		formats := []string{
			"2006-01-02",
			"Jan 2, 2006",
			"15:04",
		}

		for _, format := range formats {
			t, err := time.Parse(format, vals[0])
			if err != nil {
				continue
			}

			return t, nil
		}

		return nil, fmt.Errorf("could not decode input %v", vals[0])
	}, time.Time{})

	return dec
}
