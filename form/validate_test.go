package form_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/leapkit/core/form"
	"github.com/leapkit/core/form/validate"
)

func TestValidate(t *testing.T) {
	reqFromParams := func(params url.Values) *http.Request {
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(params.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.ParseForm()

		return req
	}

	t.Run("Valid simple request", func(t *testing.T) {
		req := reqFromParams(url.Values{
			"name": {"John"},
		})

		rules := validate.Form(
			validate.Field("name", validate.Required()),
		)

		errs := form.Validate(req, rules)
		if len(errs) > 0 {
			t.Fatalf("expected no errors, got %v", errs)
		}
	})

	t.Run("Invalid simple request", func(t *testing.T) {
		req := reqFromParams(url.Values{
			"name": {""},
		})

		rules := validate.Form(
			validate.Field("name", validate.Required()),
		)

		errs := form.Validate(req, rules)
		if len(errs) == 0 {
			t.Fatalf("expected errors, got none")
		}
	})
}
