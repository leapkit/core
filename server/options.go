package server

// Allows to specify options to the server.
type Option func(*Instance)

// WithPort Allows to specify a port for the server to listen on.
func WithPort(port string) Option {
	return func(s *Instance) {
		s.port = port
	}
}

// WithHost Allows to specify a host for the server to listen on.
func WithHost(host string) Option {
	return func(s *Instance) {
		s.host = host
	}
}
