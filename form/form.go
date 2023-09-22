package form

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/form/v4"
)

// use a single instance of Decoder, it caches struct info
var (
	decoder = form.NewDecoder()
)

func parseFormForType(r *http.Request, dst interface{}) error {
	//MultipartForm
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return err
		}

		return nil
	}

	err := r.ParseForm()
	if err != nil {
		return err
	}

	return nil
}

func Decode(r *http.Request, dst interface{}) error {
	if err := parseFormForType(r, dst); err != nil {
		return err
	}

	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
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

	return decoder.Decode(dst, r.Form)
}
