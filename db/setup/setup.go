package setup

// Manager is the interface that wraps the basic methods to
// create and drop a database.
type Manager interface {
	Create(url string) error
	Drop(url string) error
}
