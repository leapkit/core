package db

// connectionOptions for the database
type connectionOption func()

// WithDriver allows to specify the driver to use driver defaults to
// sqlite3.
func WithDriver(name string) connectionOption {
	return func() {
		driverName = name
	}
}
