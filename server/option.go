package server

type Option func(*mux)

func WithHost(host string) Option {
	return func(s *mux) {
		s.host = host
	}
}

func WithPort(port string) Option {
	return func(s *mux) {
		s.port = port
	}
}
