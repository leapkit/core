package form

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/form/v4"
)

// use a single instance of Decoder, it caches struct info
var (
	decoder = form.NewDecoder()
)

func RegisterCustomTypeFunc(fn form.DecodeCustomTypeFunc, kind interface{}) {
	decoder.RegisterCustomTypeFunc(fn, kind)
}

func Decode(r *http.Request, dst interface{}) error {
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

	data := r.Form
	if len(data) == 0 {
		r.Form = r.URL.Query()
	}

	err = decoder.Decode(dst, r.Form)
	if err != nil {
		fmt.Println(err)
	}

	return err
}
