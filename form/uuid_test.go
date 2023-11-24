package form_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/leapkit/core/form"
)

func TestUUID(t *testing.T) {
	// Test that passing a uuid as a string works
	t.Run("valid uuid", func(t *testing.T) {
		vals := url.Values{
			"id": []string{"6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		}

		tr, err := http.NewRequest("GET", "/?"+vals.Encode(), nil)
		if err != nil {
			t.Fatal(err)
		}

		tr.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		test := struct {
			ID uuid.UUID `form:"id"`
		}{}

		err = form.Decode(tr, &test)
		if err != nil {
			t.Fatal(err)
		}

		if test.ID.String() != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
			t.Fatalf("expected 6ba7b810-9dad-11d1-80b4-00c04fd430c8, got %v", test.ID.String())
		}
	})

	t.Run("invalid uuid", func(t *testing.T) {
		vals := url.Values{
			"id": []string{"22222222"},
		}

		tr, err := http.NewRequest("GET", "/?"+vals.Encode(), nil)
		if err != nil {
			t.Fatal(err)
		}

		tr.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		test := struct {
			ID uuid.UUID `form:"id"`
		}{}

		err = form.Decode(tr, &test)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("passing zero value", func(t *testing.T) {
		vals := url.Values{
			"id": []string{"00000000-0000-0000-0000-000000000000"},
		}

		tr, err := http.NewRequest("GET", "/?"+vals.Encode(), nil)
		if err != nil {
			t.Fatal(err)
		}

		tr.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		test := struct {
			ID uuid.UUID `form:"id"`
		}{}

		err = form.Decode(tr, &test)
		if err != nil {
			t.Fatal("unexpected error, should be just Zero")
		}
	})

	t.Run("passing array", func(t *testing.T) {
		vals := url.Values{
			"id": []string{"6ba7b810-9dad-11d1-80b4-00c04fd430c8", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		}

		tr, err := http.NewRequest("GET", "/?"+vals.Encode(), nil)
		if err != nil {
			t.Fatal(err)
		}

		tr.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		test := struct {
			ID []uuid.UUID `form:"id"`
		}{}

		err = form.Decode(tr, &test)
		if err != nil {
			t.Fatal(err)
		}

		if test.ID == nil {
			t.Fatalf("expected to parse the first ID")
		}

		if test.ID[0].String() != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
			t.Fatalf("expected 6ba7b810-9dad-11d1-80b4-00c04fd430c8, got %v", test.ID[0].String())
		}
	})
}

func TestDecodeUUIDSlice(t *testing.T) {
	t.Run("multiple ids", func(t *testing.T) {
		vals := url.Values{
			"ids": []string{"6ba7b810-9dad-11d1-80b4-00c04fd430c8", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		}

		tr, err := http.NewRequest("GET", "/?"+vals.Encode(), nil)
		if err != nil {
			t.Fatal(err)
		}

		tr.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		test := struct {
			IDs []uuid.UUID `form:"ids"`
		}{}

		err = form.Decode(tr, &test)
		if err != nil {
			t.Fatal(err)
		}

		if test.IDs[0].String() != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
			t.Fatalf("expected 6ba7b810-9dad-11d1-80b4-00c04fd430c8, got %v", test.IDs[0].String())
		}

		if test.IDs[1].String() != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
			t.Fatalf("expected 6ba7b810-9dad-11d1-80b4-00c04fd430c8, got %v", test.IDs[1].String())
		}
	})

	t.Run("single id", func(t *testing.T) {
		vals := url.Values{
			"ids": []string{"6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		}

		tr, err := http.NewRequest("GET", "/?"+vals.Encode(), nil)
		if err != nil {
			t.Fatal(err)
		}

		tr.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		test := struct {
			IDs []uuid.UUID `form:"ids"`
		}{}

		err = form.Decode(tr, &test)
		if err != nil {
			t.Fatal(err)
		}

		if test.IDs[0].String() != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
			t.Fatalf("expected 6ba7b810-9dad-11d1-80b4-00c04fd430c8, got %v", test.IDs[0].String())
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		vals := url.Values{
			"ids": []string{"22222222"},
		}

		tr, err := http.NewRequest("GET", "/?"+vals.Encode(), nil)
		if err != nil {
			t.Fatal(err)
		}

		tr.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		test := struct {
			IDs []uuid.UUID `form:"ids"`
		}{}

		err = form.Decode(tr, &test)
		if err == nil {
			t.Fatal("expected error")
		}
	})

}
