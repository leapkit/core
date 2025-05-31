package form_test

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"go.leapkit.dev/core/form"
)

func FuzzDecode_AllTypes(f *testing.F) {
	type AllTypes struct {
		Bool       *bool
		Int        *int
		Int8       *int8
		Int16      *int16
		Int32      *int32
		Int64      *int64
		Uint       *uint
		Uint8      *uint8
		Uint16     *uint16
		Uint32     *uint32
		Uint64     *uint64
		Float32    *float32
		Float64    *float64
		Complex64  *complex64
		Complex128 *complex128
		String     *string
		Time       time.Time
	}

	f.Add(
		true,
		int(123),
		int8(124),
		int16(125),
		int32(126),
		int64(127),
		uint(128),
		uint8(129),
		uint16(130),
		uint32(131),
		uint64(132),
		float32(25.678),
		float64(99.99),
		"hello",
	)

	f.Fuzz(func(t *testing.T, b bool, i int, i8 int8, i16 int16, i32 int32, i64 int64, ui uint, ui8 uint8, ui16 uint16, ui32 uint32, ui64 uint64, f32 float32, f64 float64, str string) {
		seed := url.Values{
			"Bool": {fmt.Sprint(b)},

			"Int":   {fmt.Sprint(i)},
			"Int8":  {fmt.Sprint(i8)},
			"Int16": {fmt.Sprint(i16)},
			"Int32": {fmt.Sprint(i32)},
			"Int64": {fmt.Sprint(i64)},

			"Uint":   {fmt.Sprint(ui)},
			"Uint8":  {fmt.Sprint(ui8)},
			"Uint16": {fmt.Sprint(ui16)},
			"Uint32": {fmt.Sprint(ui32)},
			"Uint64": {fmt.Sprint(ui64)},

			"Float32":    {fmt.Sprint(f32)},
			"Float64":    {fmt.Sprint(f64)},
			"Complex64":  {fmt.Sprint(complex64(complex(f64, float64(f32))))},
			"Complex128": {fmt.Sprint(complex64(complex(float64(f32), f64)))},
			"String":     {str},
		}

		req, err := http.NewRequest("POST", "/", strings.NewReader(seed.Encode()))
		if err != nil {
			t.Skip()
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		var s AllTypes
		err = form.Decode(req, &s)
		if err != nil {
			t.Errorf("Expected nil, got error: %v", err)
		}
		if s.Bool == nil {
			t.Error("Expected true of false, got nil")
		}
		if s.Int == nil {
			t.Errorf("Expected value %s, got nil", seed["Int"][0])
		}
		if s.Int8 == nil {
			t.Errorf("Expected value %s, got nil", seed["Int8"][0])
		}
		if s.Int16 == nil {
			t.Errorf("Expected value %s, got nil", seed["Int16"][0])
		}
		if s.Int32 == nil {
			t.Errorf("Expected value %s, got nil", seed["Int32"][0])
		}
		if s.Int64 == nil {
			t.Errorf("Expected value %s, got nil", seed["Int64"][0])
		}
		if s.Uint == nil {
			t.Errorf("Expected value %s, got nil", seed["Uint"][0])
		}
		if s.Uint8 == nil {
			t.Errorf("Expected value %s, got nil", seed["Uint8"][0])
		}
		if s.Uint16 == nil {
			t.Errorf("Expected value %s, got nil", seed["Uint16"][0])
		}
		if s.Uint32 == nil {
			t.Errorf("Expected value %s, got nil", seed["Uint32"][0])
		}
		if s.Uint64 == nil {
			t.Errorf("Expected value %s, got nil", seed["Uint64"][0])
		}
		if s.Float32 == nil {
			t.Errorf("Expected value %s, got nil", seed["Float32"][0])
		}
		if s.Float64 == nil {
			t.Errorf("Expected value %s, got nil", seed["Float64"][0])
		}
		if s.Complex64 == nil {
			t.Errorf("Expected value %s, got nil", seed["Complex64"][0])
		}
		if s.Complex128 == nil {
			t.Errorf("Expected value %s, got nil", seed["Complex128"][0])
		}
		if s.String == nil {
			t.Errorf("Expected value %s, got nil", seed["String"][0])
		}
	})
}
