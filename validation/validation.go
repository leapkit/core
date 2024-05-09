package validation

import "net/url"

// New validation with the passed field and rules.
func New(field string, rules ...Rule) validation {
	return validation{
		Field: field,
		Rules: rules,
	}
}

// validation is a struct that contains a set of rules
// that form values must comply with for a specific field.
type validation struct {
	Field string
	Rules []Rule
}

type Validations []validation

// Validate is the main method we will use to perform the validations on a form.
func (v Validations) Validate(form url.Values) Errors {
	verrs := make(map[string][]error)
	for _, validation := range v {
		for _, rule := range validation.Rules {
			err := rule(form[validation.Field])
			if err == nil {
				continue
			}

			verrs[validation.Field] = append(verrs[validation.Field], err)
		}
	}

	return verrs
}

type Errors map[string][]error
