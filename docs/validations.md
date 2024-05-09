---
title: Form Validation
index: 3
---

Leapkit provides a `validation` package that offers a flexible and reusable way to validate form data
by defining a set of validation rules that can be applied to form fields.

### How to Use

In your code, define a `validation.Validations` variable with a set of validations.
Then, call the `Validations.Validate()` method, which receives the `url.Values` form and returns
a `map[string][]error` map with validation errors per form field. Example:

```go
form := url.Values{}

validations := validation.Validations{
	validation.Validation{
		Field: "my_awesome_field",
		Rules: []Rule{
			validations.Required("my custom error"),
		},
	},
}

verrs := validations.Validate(form)
if len(verrs) > 0 {
	 handle invalid form
}
```

### Built-in Rules

You can build your set of rules for each validation by using the package's built-in functions.

```go
// General Rules:
func Required(message ...string) Rule

// String Rules:
func Matches(field string, message ...string) Rule
func MatchRegex(re \*regexp.Regexp, message ...string) Rule
func MinLength(min int, message ...string) Rule
func MaxLength(max int, message ...string) Rule
func WithinOptions(options []string, message ...string) Rule

// Number Rules:
func LessThan(value float64, message ...string) Rule
func LessThanOrEqualTo(value float64, message ...string) Rule
func GreaterThan(value float64, message ...string) Rule
func GreaterThanOrEqualTo(value float64, message ...string) Rule

// UUID Rule:
func ValidUUID(message ...string) Rule

// Time Rules:
func TimeEqualTo(u time.Time, message ...string) Rule
func TimeBefore(u time.Time, message ...string) Rule
func TimeBeforeOrEqualTo(u time.Time, message ...string) Rule
func TimeAfter(u time.Time, message ...string) Rule
func TimeAfterOrEqualTo(u time.Time, message ...string) Rule
```

### Custom validation Rules:

Alternatively, you can create your own validation rules. Example:

```go
var (
	DB *sqlx.DB
)

func IsUnique() validations.Rule {
   	return func(emails []string) error {
  		stmt, err := DB.Prepare("SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)")
  		if err != nil {
 			return err
  		}

  		for _, email := range emails {
 			var exists bool
 			if err := stmt.QueryRow(email).Scan(&exists); err != nil {
				return err
 			}

 			if exists {
				return fmt.Errorf("email '%s' already exists.", email)
 			}
  		}

  		return nil
   	}
}

// ...

validations := validation.Validations{
   	validation.Validation{
  		Field: "my_awesome_field",
  		Rules: []Rule{
 			IsUnique(),
  		},
   	},
}
```
