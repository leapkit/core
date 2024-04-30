package validation

import (
	"net/url"
	"regexp"
	"slices"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	form := url.Values{
		"first_name": []string{""},
		"last_name":  []string{""},
	}

	verrs := New(form)

	if len(verrs) != 0 {
		t.Fatalf("validation should not contains any error")
	}
}

func TestRule(test *testing.T) {
	test.Run("if value is invalid verrs should not be empty", func(t *testing.T) {
		form := url.Values{
			"input_field":  []string{""},
			"number_input": []string{"50"},
			"multiple":     []string{"one", "two", "three", "invalid_value"},
			"text_input":   []string{"lorem ipsum"},
		}

		verrs := New(form,
			Rule{
				Field: "input_field",
				Validations: []Validation{
					Required(),
				},
			},
			Rule{
				Field: "number_input",
				Validations: []Validation{
					LessThan(20),
				},
			},
			Rule{
				Field: "multiple",
				Validations: []Validation{
					WithinOptions([]string{"one", "two", "tree"}),
				},
			},
			Rule{
				Field: "text_input",
				Validations: []Validation{
					MinLength(20),
				},
			},
		)

		if len(verrs) != 4 {
			t.Fatalf("verrs should have four errors, verrs=%v", verrs)

		}
	})

	test.Run("if SkipIf is true form values should not be validated", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{""},
		}

		customCondition := true

		verrs := New(form, Rule{
			Field:  "input_field",
			SkipIf: customCondition,
			Validations: []Validation{
				Required(),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("if SkipIf is true form values should not be validated", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{""},
		}

		customCondition := true

		verrs := New(form, Rule{
			Field:  "input_field",
			SkipIf: customCondition,
			Validations: []Validation{
				Required(),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})
}

func TestValidationRequired(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"foo"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				Required(),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have an error, verrs=%v", verrs)
		}
	})

	test.Run("happy path with custom error", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{""},
		}

		customError := "My custom error"

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				Required(customError),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have an error, verrs=%v", verrs)
		}

		if !slices.Contains(verrs["input_field"], customError) {
			t.Fatalf("verrs should contain '%s' custom error", customError)
		}
	})

	test.Run("field is not in form", func(t *testing.T) {
		form := url.Values{}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				Required(),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have an error, verrs=%v", verrs)
		}
	})

	test.Run("field is empty", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{""},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				Required(),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have an error, verrs=%v", verrs)
		}
	})

	test.Run("at least a field is empty", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"one", "", "three"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				Required(),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have an error, verrs=%v", verrs)
		}
	})
}

func TestValidationMatch(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"foo"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				Match("foo"),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("happy path with custom error", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"foo"},
		}

		customError := "My custom error"

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				Match("bar", customError),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have an error, verrs=%v", verrs)
		}

		if !slices.Contains(verrs["input_field"], customError) {
			t.Fatalf("verrs should contain '%s' custom error", customError)
		}
	})
}

func TestValidationMatchRegex(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"foo"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				MatchRegex(regexp.MustCompile("f(o)+")),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}

	})

	test.Run("happy path with custom error", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"foo"},
		}

		customError := "My custom error"

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				MatchRegex(regexp.MustCompile("f(o){5,}"), customError),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if !slices.Contains(verrs["input_field"], customError) {
			t.Fatalf("verrs should contain '%s' custom error", customError)
		}
	})
}

func TestValidationLessThan(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"10"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				LessThan(50),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("happy path with custom error", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"10"},
		}

		customError := "My custom error"

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				LessThan(9, customError),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if !slices.Contains(verrs["input_field"], customError) {
			t.Fatalf("verrs should contain '%s' custom error", customError)
		}
	})

	test.Run("at least a value is greater than expected value", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"10", "100", "12"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				LessThan(20),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if len(verrs["input_field"]) == 0 {
			t.Fatal("verrs should contain an error for 'input_field' key")
		}
	})

	test.Run("value is not a number", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"is not a number"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				LessThan(10),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if !slices.Contains(verrs["input_field"], "is not a number") {
			t.Fatal("verrs should contain an 'is not a number' error")
		}
	})
}

func TestValidationLessThanOrEqualTo(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"10"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				LessThanOrEqualTo(10),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("happy path with custom error", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"10"},
		}

		customError := "My custom error"

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				LessThanOrEqualTo(9, customError),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if !slices.Contains(verrs["input_field"], customError) {
			t.Fatalf("verrs should contain '%s' custom error", customError)
		}
	})

	test.Run("at least a value is greater than expected value", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"10", "100", "12"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				LessThanOrEqualTo(20),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if len(verrs["input_field"]) == 0 {
			t.Fatal("verrs should contain an error for 'input_field' key")
		}
	})

	test.Run("value is not a number", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"is not a number"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				LessThanOrEqualTo(10),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if !slices.Contains(verrs["input_field"], "is not a number") {
			t.Fatal("verrs should contain an 'is not a number' error")
		}
	})
}

func TestValidationGreaterThan(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"10"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				GreaterThan(5),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("happy path with custom error", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"10"},
		}

		customError := "My custom error"

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				GreaterThan(11, customError),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if !slices.Contains(verrs["input_field"], customError) {
			t.Fatalf("verrs should contain '%s' custom error", customError)
		}
	})

	test.Run("at least a value is less than expected value", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"10", "100", "12"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				GreaterThan(10),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if len(verrs["input_field"]) == 0 {
			t.Fatal("verrs should contain an error for 'input_field' key")
		}
	})

	test.Run("value is not a number", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"is not a number"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				GreaterThan(10),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if !slices.Contains(verrs["input_field"], "is not a number") {
			t.Fatal("verrs should contain an 'is not a number' error")
		}
	})
}

func TestValidationGreaterThanOrEqualTo(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"10"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				GreaterThanOrEqualTo(10),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("happy path with custom error", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"10"},
		}

		customError := "My custom error"

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				GreaterThanOrEqualTo(11, customError),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if !slices.Contains(verrs["input_field"], customError) {
			t.Fatalf("verrs should contain '%s' custom error", customError)
		}
	})

	test.Run("at least a value is less than expected value", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"10", "100", "12"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				GreaterThanOrEqualTo(11),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if len(verrs["input_field"]) == 0 {
			t.Fatal("verrs should contain an error for 'input_field' key")
		}
	})

	test.Run("value is not a number", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"is not a number"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				GreaterThanOrEqualTo(10),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if !slices.Contains(verrs["input_field"], "is not a number") {
			t.Fatal("verrs should contain an 'is not a number' error")
		}
	})
}

func TestValidationMinLength(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"lorem ipsum"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				MinLength(3),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("happy path with custom error", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"lorem ipsum"},
		}

		customError := "My custom error"

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				MinLength(20, customError),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if !slices.Contains(verrs["input_field"], customError) {
			t.Fatalf("verrs should contain '%s' custom error", customError)
		}
	})

	test.Run("value with spaces", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"     text     "},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				MinLength(5),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})

	test.Run("value length is equal to min length should be valid", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"foo"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				MinLength(3),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("at least a value length is less than expected value", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"long text 1", "a long text 1", "short"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				MinLength(10),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if len(verrs["input_field"]) == 0 {
			t.Fatal("verrs should contain an error for 'input_field' key")
		}
	})
}

func TestValidationMaxLength(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"lorem ipsum"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				MaxLength(20),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("happy path with custom error", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"lorem ipsum"},
		}

		customError := "My custom error"

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				MaxLength(10, customError),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if !slices.Contains(verrs["input_field"], customError) {
			t.Fatalf("verrs should contain '%s' custom error", customError)
		}
	})

	test.Run("value with spaces correct", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"     long expression     "},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				MaxLength(15),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("value with spaces incorrect", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"     long expression     "},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				MaxLength(10),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})

	test.Run("value length is equal to max length should be valid", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"foo"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				MaxLength(3),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("at least a value length is greater than expected value", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"short text 1", "short text 2", "short text short text short text"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				MaxLength(15),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}

		if len(verrs["input_field"]) == 0 {
			t.Fatal("verrs should contain an error for 'input_field' key")
		}
	})
}

func TestValidationWithinOptions(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"one", "two"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				WithinOptions([]string{"one", "two", "three"}),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("at least a value is not in option list", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"one", "two", "three", "four"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				WithinOptions([]string{"one", "two", "three"}),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})
}

func TestValidaitonUUID(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"bc604324-4c14-4e17-ada9-240b27637ee5"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				ValidUUID(),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("value is not a uuid", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"invalid-uuid-4e17-ada9-240b27637ee5"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				ValidUUID(),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})
}

func TestValidationTimeEqualTo(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"2026-06-26"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeEqualTo(time.Date(2026, time.June, 26, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("value is not equal to time", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"2026-06-26"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeEqualTo(time.Date(2026, time.June, 27, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})

	test.Run("value is not a time", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"invalid value"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeEqualTo(time.Date(2026, time.June, 26, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})
}

func TestValidationTimeBefore(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"2026-06-26"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeBefore(time.Date(2026, time.June, 27, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("value is after the time", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"2026-06-27"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeBefore(time.Date(2026, time.June, 26, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})

	test.Run("value is not a time", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"invalid value"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeBefore(time.Date(2026, time.June, 26, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})
}

func TestValidationTimeBeforeOrEqualTo(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"2026-06-26"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeBeforeOrEqualTo(time.Date(2026, time.June, 27, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("value is after the time", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"2026-06-27"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeBeforeOrEqualTo(time.Date(2026, time.June, 26, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})

	test.Run("value is equal to time", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"2026-06-26"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeBeforeOrEqualTo(time.Date(2026, time.June, 26, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("value is not a time", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"invalid value"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeBeforeOrEqualTo(time.Date(2026, time.June, 26, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})
}

func TestValidationTimeAfter(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"2026-06-26"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeAfter(time.Date(2026, time.June, 25, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("value is before the time", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"2026-06-25"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeAfter(time.Date(2026, time.June, 26, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})

	test.Run("value is not a time", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"invalid value"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeAfter(time.Date(2026, time.June, 26, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})
}

func TestValidationTimeAfterOrEqualTo(test *testing.T) {
	test.Run("happy path", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"2026-06-27"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeAfterOrEqualTo(time.Date(2026, time.June, 26, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("value is before the time", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"2026-06-26"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeAfterOrEqualTo(time.Date(2026, time.June, 27, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})

	test.Run("value is equal to time", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"2026-06-26"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeAfterOrEqualTo(time.Date(2026, time.June, 26, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) != 0 {
			t.Fatalf("verrs should not have errors, verrs=%v", verrs)
		}
	})

	test.Run("value is not a time", func(t *testing.T) {
		form := url.Values{
			"input_field": []string{"invalid value"},
		}

		verrs := New(form, Rule{
			Field: "input_field",
			Validations: []Validation{
				TimeAfterOrEqualTo(time.Date(2026, time.June, 26, 0, 0, 0, 0, time.UTC)),
			},
		})

		if len(verrs) == 0 {
			t.Fatalf("verrs should have errors, verrs=%v", verrs)
		}
	})
}
