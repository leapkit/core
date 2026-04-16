package form_test

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.leapkit.dev/core/form"
)

type customStringer string

func (cs customStringer) String() string {
	return string(cs)
}

func TestDecode(t *testing.T) {
	t.Run("correct happy path", func(t *testing.T) {
		data := url.Values{}
		data.Set("Name", "Alice")
		data.Set("Age", "30")
		data.Set("Email", "alice@example.com")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Name  string
			Age   int
			Email string
		}
		if err := form.Decode(req, &s); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if s.Name != "Alice" || s.Age != 30 || s.Email != "alice@example.com" {
			t.Errorf("unexpected struct: %+v", s)
		}
	})

	t.Run("error invalid destination", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/", nil)

		var notPtr struct{}
		err := form.Decode(req, notPtr)
		if err == nil {
			t.Error("expected error for non-pointer dst")
		}

		var notStruct int
		err = form.Decode(req, &notStruct)
		if err == nil {
			t.Error("expected error for non-struct pointer dst")
		}
	})

	t.Run("correct decoding struct with form tag", func(t *testing.T) {
		data := url.Values{}
		data.Set("user", "bob")
		data.Set("is_active", "true")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Username string `form:"user"`
			Active   bool   `form:"is_active"`
		}

		err := form.Decode(req, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Username != "bob" || !s.Active {
			t.Errorf("unexpected struct: %+v", s)
		}
	})

	t.Run("correct skipping field if form tag does not match with form value", func(t *testing.T) {
		data := url.Values{}
		data.Set("Username", "bob")
		data.Set("is_active", "true")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		type tagged struct {
			Username string `form:"user"`
			Active   bool   `form:"is_active"`
		}
		var s tagged
		err := form.Decode(req, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Username == "bob" || !s.Active {
			t.Errorf("unexpected struct: %+v", s)
		}
	})

	t.Run("correct do not decode unexported field", func(t *testing.T) {
		vals := url.Values{}
		vals.Set("Name", "test")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			name string `form:"Name"`
		}

		if err := form.Decode(req, &s); err != nil {
			t.Fatal("expected error for invalid time field value")
		}

		if s.name == "test" {
			t.Error("Expected empty string, got 'test'")
		}
	})

	t.Run("correct skipping field when form tag if middle dash", func(t *testing.T) {
		data := url.Values{}
		data.Set("Username", "bob")
		data.Set("is_active", "true")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Username string `form:"-"`
			Active   bool   `form:"is_active"`
		}
		err := form.Decode(req, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Username == "bob" || !s.Active {
			t.Errorf("unexpected struct: %+v", s)
		}
	})

	t.Run("correct decoding custom type", func(t *testing.T) {
		type CustomTestInt int

		form.RegisterCustomTypeFunc(func(value string) (any, error) {
			if len(value) == 0 {
				return CustomTestInt(0), nil
			}

			var v int
			_, err := fmt.Sscanf(value, "custom-%d", &v)
			if err != nil {
				return nil, fmt.Errorf("invalid custom int: %v", err)
			}
			return CustomTestInt(v), nil
		}, CustomTestInt(0))

		data := url.Values{}
		data.Set("ID", "custom-42")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			ID CustomTestInt
		}
		err := form.Decode(req, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.ID != CustomTestInt(42) {
			t.Errorf("unexpected custom type: %+v", s)
		}
	})

	t.Run("correct decoding custom type should continue with the next fields", func(t *testing.T) {
		data := url.Values{}
		data.Set("Date", "2025-06-26")
		data.Set("ID", "2")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Date time.Time
			ID   int
		}
		err := form.Decode(req, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if s.Date.Format("2006-01-02") != "2025-06-26" || s.ID != 2 {
			t.Errorf("unexpected custom type: %+v", s)
		}
	})

	t.Run("correct decoding slice of data types", func(t *testing.T) {
		data := url.Values{}
		data.Add("Tags", "a")
		data.Add("Tags", "b")
		data.Add("Tags", "c")
		data.Add("NumSlc", "1")
		data.Add("NumSlc", "2")
		data.Add("NumSlc", "3")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Tags   []string
			NumSlc []int
		}

		err := form.Decode(req, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !slices.Equal(s.Tags, []string{"a", "b", "c"}) {
			t.Errorf("unexpected Tag slice values: %+v", s.Tags)
		}

		if !slices.Equal(s.NumSlc, []int{1, 2, 3}) {
			t.Errorf("unexpected Num slice values: %+v", s.NumSlc)
		}
	})

	t.Run("correct decoding slice of custom types", func(t *testing.T) {
		type CustomString string
		form.RegisterCustomTypeFunc(func(value string) (any, error) {
			var v string
			_, err := fmt.Sscanf(value, "custom-%s", &v)
			if err != nil {
				return nil, fmt.Errorf("invalid custom string: %v", err)
			}
			return CustomString(v), nil
		}, CustomString(""))

		formData := url.Values{}
		formData.Add("Slc", "custom-test1")
		formData.Add("Slc", "custom-test2")
		formData.Add("Slc", "custom-test3")
		req, _ := http.NewRequest("POST", "/", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Slc []CustomString
		}
		err := form.Decode(req, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []CustomString{"test1", "test2", "test3"}
		if !slices.Equal(expected, s.Slc) {
			t.Errorf("unexpected custom slice: got %+v, want %+v", s.Slc, expected)
		}
	})

	t.Run("correct decoding multipart form", func(t *testing.T) {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		w.WriteField("Name", "Eve")
		w.WriteField("Age", "22")
		w.WriteField("Email", "eve@example.com")
		w.Close()

		req, _ := http.NewRequest("POST", "/", &b)
		req.Header.Set("Content-Type", w.FormDataContentType())

		var s struct {
			Name  string
			Age   int
			Email string
		}
		err := form.Decode(req, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Name != "Eve" || s.Age != 22 || s.Email != "eve@example.com" {
			t.Errorf("unexpected struct: %+v", s)
		}
	})

	t.Run("correct decoding from URL params", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/?Name=Zed&Age=40&Email=zed@example.com", nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Name  string
			Age   int
			Email string
		}
		err := form.Decode(req, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Name != "Zed" || s.Age != 40 || s.Email != "zed@example.com" {
			t.Errorf("unexpected struct: %+v", s)
		}
	})

	t.Run("correct decoding nested struct", func(t *testing.T) {
		data := url.Values{}
		data.Set("ID", "abc123")
		data.Set("Name", "Nested")
		data.Set("Age", "50")
		data.Set("Email", "test@test.com")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		type BasicStruct struct {
			Name  string
			Age   int
			Email string
		}

		var s struct {
			ID string
			BasicStruct
		}
		err := form.Decode(req, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Name != "Nested" || s.Age != 50 || s.ID != "abc123" {
			t.Errorf("unexpected nested struct: %+v", s)
		}
	})

	t.Run("error decoding invalid data type", func(t *testing.T) {
		data := url.Values{}
		data.Set("Age", "invalid value")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Name  string
			Age   int
			Email string
		}

		if err := form.Decode(req, &s); err == nil {
			t.Error("expected error for invalid int value")
		}
	})

	t.Run("error decoding invalid custom type", func(t *testing.T) {
		type CustomTestString string
		form.RegisterCustomTypeFunc(func(value string) (any, error) {
			var v string
			_, err := fmt.Sscanf(value, "custom-%s", &v)
			if err != nil {
				return nil, fmt.Errorf("invalid custom string: %v", err)
			}
			return CustomTestString(v), nil
		}, CustomTestString(""))

		data := url.Values{}
		data.Set("ID", "bad-value")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			ID CustomTestString
		}

		if err := form.Decode(req, &s); err == nil {
			t.Error("expected error for invalid custom type value")
		}
	})

	t.Run("decoding all builtin types", func(t *testing.T) {
		type AllTypes struct {
			Bool    bool
			Int     int
			Int8    int8
			Int16   int16
			Int32   int32
			Int64   int64
			Uint    uint
			Uint8   uint8
			Uint16  uint16
			Uint32  uint32
			Uint64  uint64
			Float32 float32
			Float64 float64
			String  string
			Time    time.Time
		}

		valid := url.Values{
			"Bool":    {"true"},
			"Int":     {"-42"},
			"Int8":    {"-8"},
			"Int16":   {"-16000"},
			"Int32":   {"-32000"},
			"Int64":   {"-64000"},
			"Uint":    {"42"},
			"Uint8":   {"8"},
			"Uint16":  {"16000"},
			"Uint32":  {"32000"},
			"Uint64":  {"64000"},
			"Float32": {"3.14"},
			"Float64": {"2.71828"},
			"String":  {"hello"},
			"Time":    {"2025-06-26"},
		}

		t.Run("all built-in types valid", func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/", strings.NewReader(valid.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			var s AllTypes
			err := form.Decode(req, &s)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !s.Bool ||
				s.Int != -42 ||
				s.Int8 != -8 ||
				s.Int16 != -16000 ||
				s.Int32 != -32000 ||
				s.Int64 != -64000 ||
				s.Uint != 42 ||
				s.Uint8 != 8 ||
				s.Uint16 != 16000 ||
				s.Uint32 != 32000 ||
				s.Uint64 != 64000 ||
				s.Float32 != float32(3.14) ||
				s.Float64 != 2.71828 ||
				s.String != "hello" ||
				s.Time.Format("2006-01-02") != "2025-06-26" {
				t.Errorf("unexpected struct: %+v", s)
			}
		})

		invalidCases := []struct {
			field string
			val   string
		}{
			{"Bool", "notABool"},
			{"Int", "notAnInt"},
			{"Int8", "notAnInt8"},
			{"Int16", "notAnInt16"},
			{"Int32", "notAnInt32"},
			{"Int64", "notAnInt64"},
			{"Uint", "-1"},
			{"Uint8", "-2"},
			{"Uint16", "-3"},
			{"Uint32", "-4"},
			{"Uint64", "-5"},
			{"Float32", "notaFloat"},
			{"Float64", "notaFloat"},
		}

		for _, tc := range invalidCases {
			t.Run("invalid "+tc.field, func(t *testing.T) {
				vals := url.Values{}
				vals.Set(tc.field, tc.val)
				req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				var s AllTypes
				err := form.Decode(req, &s)
				if err == nil {
					t.Errorf("expected error for field %s with value %q", tc.field, tc.val)
				}
			})
		}
	})

	t.Run("correct decoding slice of custom type", func(t *testing.T) {
		type Item struct {
			ID    int
			Label string
			Flag  bool
		}

		type Payload struct {
			Items []Item
		}

		formData := url.Values{}
		formData.Set("Items[0].ID", "1")
		formData.Set("Items[0].Label", "foo")
		formData.Set("Items[0].Flag", "true")
		formData.Set("Items[1].ID", "2")
		formData.Set("Items[1].Label", "bar")
		formData.Set("Items[1].Flag", "false")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var p Payload
		err := form.Decode(req, &p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []Item{
			{ID: 1, Label: "foo", Flag: true},
			{ID: 2, Label: "bar", Flag: false},
		}

		if !slices.Equal(p.Items, expected) {
			t.Errorf("unexpected decoded slice of structs: got %+v, want %+v", p.Items, expected)
		}
	})

	t.Run("Decoding struct with pointers", func(t *testing.T) {
		type PtrStruct struct {
			Name  *string
			Age   *int
			Score *float64
			Flag  *bool
		}
		vals := url.Values{}
		vals.Set("Name", "Pointer")
		vals.Set("Age", "33")
		vals.Set("Score", "99.5")
		vals.Set("Flag", "true")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s PtrStruct
		err := form.Decode(req, &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Name == nil || *s.Name != "Pointer" {
			t.Errorf("expected Name pointer to 'Pointer', got %+v", s.Name)
		}
		if s.Age == nil || *s.Age != 33 {
			t.Errorf("expected Age pointer to 33, got %+v", s.Age)
		}
		if s.Score == nil || *s.Score != 99.5 {
			t.Errorf("expected Score pointer to 99.5, got %+v", s.Score)
		}
		if s.Flag == nil || *s.Flag != true {
			t.Errorf("expected Flag pointer to true, got %+v", s.Flag)
		}

		// Test nil pointers when fields are missing
		req2, _ := http.NewRequest("POST", "/", strings.NewReader(""))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		var s2 PtrStruct
		err = form.Decode(req2, &s2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s2.Name != nil || s2.Age != nil || s2.Score != nil || s2.Flag != nil {
			t.Errorf("expected all pointers to be nil, got %+v", s2)
		}
	})

	t.Run("decoding pointer slices", func(t *testing.T) {
		type Item struct {
			ID   *int
			Name *string
		}
		type Payload struct {
			Items []*Item
		}
		vals := url.Values{}
		vals.Set("Items[0].ID", "1")
		vals.Set("Items[0].Name", "foo")
		vals.Set("Items[1].ID", "2")
		vals.Set("Items[1].Name", "bar")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var p Payload
		err := form.Decode(req, &p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(p.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(p.Items))
		}
		if p.Items[0] == nil || p.Items[0].ID == nil || *p.Items[0].ID != 1 || p.Items[0].Name == nil || *p.Items[0].Name != "foo" {
			t.Errorf("unexpected first item: %+v", p.Items[0])
		}
		if p.Items[1] == nil || p.Items[1].ID == nil || *p.Items[1].ID != 2 || p.Items[1].Name == nil || *p.Items[1].Name != "bar" {
			t.Errorf("unexpected second item: %+v", p.Items[1])
		}
	})

	t.Run("decoding pointer struct", func(t *testing.T) {
		vals := url.Values{}
		vals.Set("Bs.Name", "Test")
		vals.Set("Bs.Age", "2")
		vals.Set("Address", "Test street")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		type BasicStruct struct {
			Name string
			Age  int
		}

		var s struct {
			Bs      *BasicStruct
			Address string
		}

		if err := form.Decode(req, &s); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if s.Bs.Age != 2 || s.Bs.Name != "Test" || s.Address != "Test street" {
			t.Errorf("unexpected struct: %+v", s.Bs)
		}
	})

	t.Run("decoding anonymous pointer struct", func(t *testing.T) {
		vals := url.Values{}
		vals.Set("Name", "Test")
		vals.Set("Age", "2")
		vals.Set("Address", "Test street")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		type BasicStruct struct {
			Name string
			Age  int
		}

		var s struct {
			*BasicStruct
			Address string
		}

		if err := form.Decode(req, &s); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if s.Age != 2 || s.Name != "Test" || s.Address != "Test street" {
			t.Errorf("unexpected struct: %+v", s)
		}
	})

	t.Run("decoding slice of pointer type", func(t *testing.T) {
		vals := url.Values{}
		vals.Add("Slc", "Test1")
		vals.Add("Slc", "Test2")
		vals.Add("Slc", "Test3")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Slc []*string
		}

		if err := form.Decode(req, &s); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		value1 := "Test1"
		value2 := "Test2"
		value3 := "Test3"

		areEqual := slices.EqualFunc(s.Slc, []*string{&value1, &value2, &value3}, func(a, b *string) bool {
			return *a == *b
		})

		if !areEqual {
			var current []string
			for _, val := range s.Slc {
				current = append(current, *val)
			}

			t.Errorf("unexpected struct: %+v", current)
		}
	})

	t.Run("incorrect decoding slice of pointer type, invalid value", func(t *testing.T) {
		vals := url.Values{}
		vals.Add("Slc", "1")
		vals.Add("Slc", "2")
		vals.Add("Slc", "3")
		vals.Add("Slc", "invalid")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Slc []*int
		}

		if err := form.Decode(req, &s); err == nil {
			t.Fatal("expected error for invalid value")
		}
	})

	t.Run("decoding skipping unexported field", func(t *testing.T) {
		vals := url.Values{}
		vals.Set("name", "Test")
		vals.Add("Age", "50")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			name string
			Age  int
		}

		if err := form.Decode(req, &s); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if s.name == "Test" {
			t.Errorf("unexpected name field value: %s", s.name)
		}

		if s.Age != 50 {
			t.Errorf("unexpected Age field value: %d", s.Age)
		}
	})

	t.Run("incorrect decoding invalid struct field value", func(t *testing.T) {
		vals := url.Values{}
		vals.Set("Person.Name", "Test")
		vals.Add("Person.Age", "invalid")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Person struct {
				Name string
				Age  int
			}
		}

		if err := form.Decode(req, &s); err == nil {
			t.Fatal("expected error for invalid struct field value")
		}
	})

	t.Run("incorrect decoding invalid slice field value", func(t *testing.T) {
		vals := url.Values{}
		vals.Add("Slc", "1")
		vals.Add("Slc", "2")
		vals.Add("Slc", "Invalid value")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Slc []int
		}

		if err := form.Decode(req, &s); err == nil {
			t.Fatal("expected error for invalid struct field value")
		}
	})

	t.Run("incorrect decoding invalid slice struct field value", func(t *testing.T) {
		vals := url.Values{}
		vals.Set("Slc[0].Value", "1")
		vals.Set("Slc[1].Value", "2")
		vals.Set("Slc[2].Value", "Invalid value")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Slc []struct {
				Value int
			}
		}

		if err := form.Decode(req, &s); err == nil {
			t.Fatal("expected error for invalid struct field value")
		}
	})
	t.Run("correct skipping decoding invalid slice field value when form does not have a value", func(t *testing.T) {
		vals := url.Values{}
		vals.Set("Name", "test")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Name string
			Slc  []int
		}

		if err := form.Decode(req, &s); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if s.Name != "test" {
			t.Errorf("unexpected Name field value: %s", s.Name)
		}

		if len(s.Slc) > 0 {
			t.Errorf("expected to be empty, got: %+v", s.Slc)
		}
	})

	t.Run("correct skipping time field value when form does not have a value", func(t *testing.T) {
		vals := url.Values{}
		vals.Set("Name", "test")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Name string
			Time time.Time
		}

		if err := form.Decode(req, &s); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if s.Name != "test" {
			t.Errorf("unexpected Name field value: %s", s.Name)
		}

		if !s.Time.IsZero() {
			t.Errorf("expected to be empty, got: %+v", s.Time)
		}
	})

	t.Run("incorrect invalid time field value", func(t *testing.T) {
		vals := url.Values{}
		vals.Set("Name", "test")
		vals.Set("Time", "invalid")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Name string
			Time time.Time
		}

		if err := form.Decode(req, &s); err == nil {
			t.Fatal("expected error for invalid time field value")
		}
	})

	t.Run("correct RegisterCustomTypeFunc with concrete type", func(t *testing.T) {
		type CustomType1 int

		form.RegisterCustomTypeFunc(func(s string) (CustomType1, error) {
			n, err := strconv.Atoi(s)
			return CustomType1(n), err
		})

		vals := url.Values{}
		vals.Set("field_a", "42")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			FieldA CustomType1 `form:"field_a"`
		}

		if err := form.Decode(req, &s); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if s.FieldA != 42 {
			t.Errorf("expected FieldA to be 42, got %v", s.FieldA)
		}
	})

	t.Run("correct RegisterCustomTypeFunc with pointer field", func(t *testing.T) {
		type CustomType1 int

		form.RegisterCustomTypeFunc(func(s string) (CustomType1, error) {
			n, err := strconv.Atoi(s)
			return CustomType1(n), err
		})

		vals := url.Values{}
		vals.Set("present", "7")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Present *CustomType1 `form:"present"`
			Absent  *CustomType1 `form:"absent"`
		}

		if err := form.Decode(req, &s); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if s.Present == nil || *s.Present != 7 {
			t.Errorf("expected Present to be pointer to 7, got %v", s.Present)
		}

		if s.Absent != nil {
			t.Errorf("expected Absent to be nil, got %v", s.Absent)
		}
	})

	t.Run("correct RegisterCustomTypeFunc with any and customType hint", func(t *testing.T) {
		type CustomType2 struct{ V int }

		form.RegisterCustomTypeFunc(func(s string) (any, error) {
			n, err := strconv.Atoi(s)
			return CustomType2{V: n}, err
		}, CustomType2{})

		vals := url.Values{}
		vals.Set("field", "10")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Field CustomType2 `form:"field"`
		}

		if err := form.Decode(req, &s); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if s.Field.V != 10 {
			t.Errorf("expected Field.V to be 10, got %v", s.Field.V)
		}
	})

	t.Run("correct RegisterCustomTypeFunc with interface type", func(t *testing.T) {
		form.RegisterCustomTypeFunc(func(s string) (fmt.Stringer, error) {
			return customStringer(s), nil
		})

		vals := url.Values{}
		vals.Set("field", "hello")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Field fmt.Stringer `form:"field"`
		}

		if err := form.Decode(req, &s); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if s.Field == nil {
			t.Fatal("expected Field to not be nil")
		}

		if s.Field.String() != "hello" {
			t.Errorf("expected Field.String() to be %q, got %q", "hello", s.Field.String())
		}
	})

	t.Run("incorrect RegisterCustomTypeFunc decoder error", func(t *testing.T) {
		type CustomType3 int

		form.RegisterCustomTypeFunc(func(s string) (CustomType3, error) {
			return 0, fmt.Errorf("custom decode error: %q", s)
		})

		vals := url.Values{}
		vals.Set("field", "bad")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var s struct {
			Field CustomType3 `form:"field"`
		}

		err := form.Decode(req, &s)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "custom decode error") {
			t.Errorf("expected error to contain %q, got %q", "custom decode error", err.Error())
		}
	})

	t.Run("correct decoding self-referencing struct with data", func(t *testing.T) {
		type Node struct {
			ID       string
			Name     string
			Parent   *Node
			Sibling  *Node
			Metadata map[string]string
		}

		vals := url.Values{}
		vals.Set("ID", "node-1")
		vals.Set("Name", "Root Node")
		vals.Set("Parent.ID", "parent-1")
		vals.Set("Parent.Name", "Parent Node")
		vals.Set("Sibling.ID", "sibling-1")
		vals.Set("Sibling.Name", "Sibling Node")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var node Node
		if err := form.Decode(req, &node); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if node.ID != "node-1" {
			t.Errorf("expected ID to be %q, got %q", "node-1", node.ID)
		}

		if node.Name != "Root Node" {
			t.Errorf("expected Name to be %q, got %q", "Root Node", node.Name)
		}

		if node.Parent == nil {
			t.Fatal("expected Parent to not be nil")
		}

		if node.Parent.ID != "parent-1" {
			t.Errorf("expected Parent.ID to be %q, got %q", "parent-1", node.Parent.ID)
		}

		if node.Parent.Name != "Parent Node" {
			t.Errorf("expected Parent.Name to be %q, got %q", "Parent Node", node.Parent.Name)
		}

		if node.Sibling == nil {
			t.Fatal("expected Sibling to not be nil")
		}

		if node.Sibling.ID != "sibling-1" {
			t.Errorf("expected Sibling.ID to be %q, got %q", "sibling-1", node.Sibling.ID)
		}

		if node.Sibling.Name != "Sibling Node" {
			t.Errorf("expected Sibling.Name to be %q, got %q", "Sibling Node", node.Sibling.Name)
		}

		// Verify that deeply nested fields don't get populated (no infinite recursion)
		if node.Parent.Parent != nil {
			t.Error("expected Parent.Parent to be nil (no form data)")
		}

		if node.Sibling.Sibling != nil {
			t.Error("expected Sibling.Sibling to be nil (no form data)")
		}
	})

	t.Run("correct skipping self-referencing struct without data", func(t *testing.T) {
		type Node struct {
			ID     string
			Name   string
			Parent *Node
		}

		vals := url.Values{}
		vals.Set("ID", "node-1")
		vals.Set("Name", "Root Node")
		// No Parent.* fields

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var node Node
		if err := form.Decode(req, &node); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if node.ID != "node-1" {
			t.Errorf("expected ID to be %q, got %q", "node-1", node.ID)
		}

		if node.Name != "Root Node" {
			t.Errorf("expected Name to be %q, got %q", "Root Node", node.Name)
		}

		// Parent should be nil since no form data exists for it
		if node.Parent != nil {
			t.Errorf("expected Parent to be nil, got %+v", node.Parent)
		}
	})

	t.Run("correct decoding multiple levels of nested self-referencing structs", func(t *testing.T) {
		type TreeNode struct {
			ID       string
			Value    int
			Parent   *TreeNode
			Children []*TreeNode
		}

		vals := url.Values{}
		// Root node
		vals.Set("ID", "root")
		vals.Set("Value", "100")

		// Parent of root
		vals.Set("Parent.ID", "grandparent")
		vals.Set("Parent.Value", "200")

		// Grandparent of root
		vals.Set("Parent.Parent.ID", "great-grandparent")
		vals.Set("Parent.Parent.Value", "300")

		// Children of root
		vals.Set("Children[0].ID", "child-1")
		vals.Set("Children[0].Value", "10")
		vals.Set("Children[1].ID", "child-2")
		vals.Set("Children[1].Value", "20")

		// Grandchildren through first child
		vals.Set("Children[0].Children[0].ID", "grandchild-1")
		vals.Set("Children[0].Children[0].Value", "5")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var root TreeNode
		if err := form.Decode(req, &root); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify root
		if root.ID != "root" || root.Value != 100 {
			t.Errorf("expected root (root, 100), got (%s, %d)", root.ID, root.Value)
		}

		// Verify parent
		if root.Parent == nil {
			t.Fatal("expected Parent to not be nil")
		}
		if root.Parent.ID != "grandparent" || root.Parent.Value != 200 {
			t.Errorf("expected parent (grandparent, 200), got (%s, %d)", root.Parent.ID, root.Parent.Value)
		}

		// Verify grandparent
		if root.Parent.Parent == nil {
			t.Fatal("expected Parent.Parent to not be nil")
		}
		if root.Parent.Parent.ID != "great-grandparent" || root.Parent.Parent.Value != 300 {
			t.Errorf("expected grandparent (great-grandparent, 300), got (%s, %d)",
				root.Parent.Parent.ID, root.Parent.Parent.Value)
		}

		// Verify great-grandparent has no parent (no form data)
		if root.Parent.Parent.Parent != nil {
			t.Error("expected great-grandparent's parent to be nil")
		}

		// Verify children
		if len(root.Children) != 2 {
			t.Fatalf("expected 2 children, got %d", len(root.Children))
		}
		if root.Children[0].ID != "child-1" || root.Children[0].Value != 10 {
			t.Errorf("expected child-1 (child-1, 10), got (%s, %d)", root.Children[0].ID, root.Children[0].Value)
		}
		if root.Children[1].ID != "child-2" || root.Children[1].Value != 20 {
			t.Errorf("expected child-2 (child-2, 20), got (%s, %d)", root.Children[1].ID, root.Children[1].Value)
		}

		// Verify grandchildren
		if len(root.Children[0].Children) != 1 {
			t.Fatalf("expected 1 grandchild, got %d", len(root.Children[0].Children))
		}
		if root.Children[0].Children[0].ID != "grandchild-1" || root.Children[0].Children[0].Value != 5 {
			t.Errorf("expected grandchild (grandchild-1, 5), got (%s, %d)",
				root.Children[0].Children[0].ID, root.Children[0].Children[0].Value)
		}

		// Verify second child has no children (no form data)
		if len(root.Children[1].Children) > 0 {
			t.Errorf("expected child-2 to have no children, got %d", len(root.Children[1].Children))
		}
	})

	t.Run("correct decoding partial nested data with missing fields", func(t *testing.T) {
		type Person struct {
			ID        string
			FirstName string
			LastName  string
			Age       int
			Spouse    *Person
			Manager   *Person
		}

		vals := url.Values{}
		// Main person - all fields
		vals.Set("ID", "person-1")
		vals.Set("FirstName", "John")
		vals.Set("LastName", "Doe")
		vals.Set("Age", "30")

		// Spouse - only ID and FirstName
		vals.Set("Spouse.ID", "person-2")
		vals.Set("Spouse.FirstName", "Jane")
		// Missing: Spouse.LastName, Spouse.Age

		// Manager - only ID
		vals.Set("Manager.ID", "person-3")
		// Missing: Manager.FirstName, Manager.LastName, Manager.Age

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var person Person
		if err := form.Decode(req, &person); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify main person
		if person.ID != "person-1" {
			t.Errorf("expected ID %q, got %q", "person-1", person.ID)
		}
		if person.FirstName != "John" {
			t.Errorf("expected FirstName %q, got %q", "John", person.FirstName)
		}
		if person.LastName != "Doe" {
			t.Errorf("expected LastName %q, got %q", "Doe", person.LastName)
		}
		if person.Age != 30 {
			t.Errorf("expected Age %d, got %d", 30, person.Age)
		}

		// Verify spouse partial data
		if person.Spouse == nil {
			t.Fatal("expected Spouse to not be nil")
		}
		if person.Spouse.ID != "person-2" {
			t.Errorf("expected Spouse.ID %q, got %q", "person-2", person.Spouse.ID)
		}
		if person.Spouse.FirstName != "Jane" {
			t.Errorf("expected Spouse.FirstName %q, got %q", "Jane", person.Spouse.FirstName)
		}
		// Missing fields should be zero values
		if person.Spouse.LastName != "" {
			t.Errorf("expected Spouse.LastName to be empty, got %q", person.Spouse.LastName)
		}
		if person.Spouse.Age != 0 {
			t.Errorf("expected Spouse.Age to be 0, got %d", person.Spouse.Age)
		}

		// Verify manager with only ID
		if person.Manager == nil {
			t.Fatal("expected Manager to not be nil")
		}
		if person.Manager.ID != "person-3" {
			t.Errorf("expected Manager.ID %q, got %q", "person-3", person.Manager.ID)
		}
		if person.Manager.FirstName != "" {
			t.Errorf("expected Manager.FirstName to be empty, got %q", person.Manager.FirstName)
		}

		// Verify nested references are nil (no form data)
		if person.Spouse.Spouse != nil {
			t.Error("expected Spouse.Spouse to be nil")
		}
		if person.Spouse.Manager != nil {
			t.Error("expected Spouse.Manager to be nil")
		}
		if person.Manager.Spouse != nil {
			t.Error("expected Manager.Spouse to be nil")
		}
		if person.Manager.Manager != nil {
			t.Error("expected Manager.Manager to be nil")
		}
	})

	t.Run("correct decoding cross-referencing structs", func(t *testing.T) {
		type LinkedNode struct {
			ID       string
			Value    string
			Next     *LinkedNode
			Previous *LinkedNode
		}

		vals := url.Values{}
		// Node A
		vals.Set("ID", "node-a")
		vals.Set("Value", "A")

		// Node A -> Next = Node B
		vals.Set("Next.ID", "node-b")
		vals.Set("Next.Value", "B")

		// Node B -> Next = Node C
		vals.Set("Next.Next.ID", "node-c")
		vals.Set("Next.Next.Value", "C")

		// Node A -> Previous = Node Z
		vals.Set("Previous.ID", "node-z")
		vals.Set("Previous.Value", "Z")

		// Node Z -> Previous = Node Y
		vals.Set("Previous.Previous.ID", "node-y")
		vals.Set("Previous.Previous.Value", "Y")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var node LinkedNode
		if err := form.Decode(req, &node); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify current node
		if node.ID != "node-a" || node.Value != "A" {
			t.Errorf("expected (node-a, A), got (%s, %s)", node.ID, node.Value)
		}

		// Verify forward chain
		if node.Next == nil {
			t.Fatal("expected Next to not be nil")
		}
		if node.Next.ID != "node-b" || node.Next.Value != "B" {
			t.Errorf("expected Next (node-b, B), got (%s, %s)", node.Next.ID, node.Next.Value)
		}

		if node.Next.Next == nil {
			t.Fatal("expected Next.Next to not be nil")
		}
		if node.Next.Next.ID != "node-c" || node.Next.Next.Value != "C" {
			t.Errorf("expected Next.Next (node-c, C), got (%s, %s)", node.Next.Next.ID, node.Next.Next.Value)
		}

		// Verify backward chain
		if node.Previous == nil {
			t.Fatal("expected Previous to not be nil")
		}
		if node.Previous.ID != "node-z" || node.Previous.Value != "Z" {
			t.Errorf("expected Previous (node-z, Z), got (%s, %s)", node.Previous.ID, node.Previous.Value)
		}

		if node.Previous.Previous == nil {
			t.Fatal("expected Previous.Previous to not be nil")
		}
		if node.Previous.Previous.ID != "node-y" || node.Previous.Previous.Value != "Y" {
			t.Errorf("expected Previous.Previous (node-y, Y), got (%s, %s)",
				node.Previous.Previous.ID, node.Previous.Previous.Value)
		}

		// Verify chains end (no more form data)
		if node.Next.Next.Next != nil {
			t.Error("expected Next.Next.Next to be nil")
		}
		if node.Previous.Previous.Previous != nil {
			t.Error("expected Previous.Previous.Previous to be nil")
		}
	})

	t.Run("correct handling empty nested struct with only nested fields", func(t *testing.T) {
		type Account struct {
			ID      string
			Balance float64
			Owner   *Account
		}

		vals := url.Values{}
		// Current account has nested owner, but no own data except nested
		vals.Set("Owner.ID", "owner-123")
		vals.Set("Owner.Balance", "1000.50")

		req, _ := http.NewRequest("POST", "/", strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var account Account
		if err := form.Decode(req, &account); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Current account should have zero values
		if account.ID != "" {
			t.Errorf("expected ID to be empty, got %q", account.ID)
		}
		if account.Balance != 0.0 {
			t.Errorf("expected Balance to be 0, got %f", account.Balance)
		}

		// Owner should be populated
		if account.Owner == nil {
			t.Fatal("expected Owner to not be nil")
		}
		if account.Owner.ID != "owner-123" {
			t.Errorf("expected Owner.ID %q, got %q", "owner-123", account.Owner.ID)
		}
		if account.Owner.Balance != 1000.50 {
			t.Errorf("expected Owner.Balance %f, got %f", 1000.50, account.Owner.Balance)
		}
	})
}
