package db

import "net/url"

// connectionOptions for the database
type connectionOption func(*connSettings)

// WithDriver allows to specify the driver to use driver defaults to
// postgres.
func WithDriver(name string) connectionOption {
	return func(cs *connSettings) {
		cs.driver = name
	}
}

// Params allows to specify additional connection parameters
// that will be encoded as URL params next to the connection string.
// params should be in key,value,key,value,... format.
// e.g Params("sslmode", "disable", "timezone", "UTC")
func Params(params ...string) connectionOption {
	vals := url.Values{}
	for i := 0; i < len(params); i += 2 {
		key := params[i]
		if i+1 >= len(params) {
			vals.Add(key, "")
			break
		}

		vals.Add(key, params[i+1])
	}

	return func(cs *connSettings) {
		cs.params = vals.Encode()
	}
}
