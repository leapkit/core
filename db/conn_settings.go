package db

import "strings"

// connSettings holds database connection settings.
// while opening the connection, its modified by
// connectionOption functions.
type connSettings struct {
	params string
	driver string
	url    string
}

func (cs connSettings) connectionURL() string {
	if cs.params == "" {
		return cs.url
	}

	if strings.Contains(cs.url, "?") {
		return cs.url + "&" + cs.params
	}

	return cs.url + "?" + cs.params
}
