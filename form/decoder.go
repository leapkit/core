package form

import (
	"net/http"
	"strings"

	"github.com/go-playground/form/v4"
)

type decoder struct {
	*form.Decoder
}

func (d *decoder) decode(r *http.Request, dst interface{}) error {
	err := d.parseForm(r)
	if err != nil {
		return err
	}

	return d.Decode(dst, r.Form)
}

func (d *decoder) parseForm(r *http.Request) error {
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

func (d *decoder) RegisterCustomTypeFunc(fn form.DecodeCustomTypeFunc, types ...interface{}) {
	d.RegisterCustomTypeFunc(fn, types...)
}
