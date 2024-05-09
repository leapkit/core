package validation

import "net/url"

// Validation is a struct that contains a set of rules
// that form values must comply with for a specific field.
type Validation struct {
	Field string
	Rules []Rule
}

type Validations []Validation

// Validate is the main method we will use to perform the validations on a form.
func (v Validations) Validate(form url.Values) map[string][]error {
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
