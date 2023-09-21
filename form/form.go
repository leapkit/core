package form

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/gofrs/uuid/v5"
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

	if err := r.ParseForm(); err != nil {
		return err
	}

	return nil
}

func Decode(r *http.Request, dst interface{}) error {
	if err := parseFormForType(r, dst); err != nil {
		return err
	}

	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		return time.Parse("2006-01-02", vals[0])
	}, time.Time{})

	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		// Attempt to parse in the format '2006-01-02'.
		t, err := time.Parse("2006-01-02", vals[0])
		if err != nil {
			// If it could not be parsed in the '2006-01-02' format, try other formats
			// as 'Jan 2, 2006' and '15:04'.
			t, err = time.Parse("Jan 2, 2006", vals[0])
			if err != nil {
				t, err = time.Parse("15:04", vals[0])
				if err != nil {
					return nil, err
				}
			}
		}

		return t, nil
	}, time.Time{})

	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		id := uuid.FromStringOrNil(vals[0])

		var nullID uuid.NullUUID

		if id != uuid.Nil {
			nullID.UUID = id
			nullID.Valid = true
		}

		return nullID, nil
	}, uuid.NullUUID{})

	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		return uuid.FromStringOrNil(vals[0]), nil
	}, uuid.UUID{})

	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		var ids []uuid.UUID

		for _, val := range vals {
			ids = append(ids, uuid.FromStringOrNil(val))
		}

		return ids, nil
	}, []uuid.UUID{})

	return decoder.Decode(dst, r.Form)
}
