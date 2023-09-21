package db

import "fmt"

type migration struct {
	Name      string
	Timestamp string
}

func (m migration) Filename() string {
	return fmt.Sprintf("%s_%s.sql", m.Timestamp, m.Name)
}
