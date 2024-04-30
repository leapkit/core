package validation

// Rule is a struct that contains a set of validations
// that form values must comply with for a specific field.
type Rule struct {
	Field       string
	SkipIf      bool
	Validations []Validation
}

func (r Rule) validate(values ...string) []string {
	verrs := []string{}

	if r.SkipIf {
		return verrs
	}

	for _, rule := range r.Validations {
		err := rule(values...)
		if err == nil {
			continue
		}

		verrs = append(verrs, err.Error())
	}

	return verrs
}
