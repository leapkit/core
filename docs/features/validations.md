---
title: Form Validation
index: 3
---

Leapkit provides the `form/validate` package that offers a flexible and reusable way to validate form data by defining a set of validation rules that can be applied to form fields.

### How to Use

Validations are a set of rules stablished for different fields passed in the request. You can define these Validations to be used in your http handlers by and call the `form.Validate` function passing the `req` (*http.Request) and handling the `validate.Errors` returned. Example:

```go
rules := validate.Fields(
	validate.Field("email", validate.Required("email is required")),
	validate.Field("password", validate.Required("password is required")),
)

verrs := form.Validate(req, rules)
if len(verrs) > 0 {
	 // handle validation errors...
}
```

### Built-in Rules

You can build your set of rules for each validation by using the package's built-in functions.

```go
// General Rules:
func Required(message ...string) Rule

// String Rules:
func Matches(field string, message ...string) Rule
func MatchRegex(re *regexp.Regexp, message ...string) Rule
func MinLength(min int, message ...string) Rule
func MaxLength(max int, message ...string) Rule
func WithinOptions(options []string, message ...string) Rule

// Number Rules:
func EqualTo(value float64, message ...string) Rule
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

### Custom validation Rules

Alternatively, you can create your own validation rules. Example:

```go
func IsUnique(db *sqlx.DB ) func([]string) error {
   	return func(emails []string) error {
     		query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	  		stmt, err := db.Prepare(query)
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
rules := validation.Fields(
	validation.New("email", validate.Required(), IsUnique(db))
)
...
```
