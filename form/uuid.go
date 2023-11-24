package form

import (
	"fmt"

	"github.com/gofrs/uuid/v5"
)

// DecodeUUID a single uuid from a string
// and returns an error if there is a problem
func DecodeUUID(vals []string) (interface{}, error) {
	uu, err := uuid.FromString(vals[0])
	if err != nil {
		err = fmt.Errorf("error parsing uuid: %w", err)
	}

	return uu, err
}

// DecodeUUIDSlice decodes a slice of uuids from a string
// and returns an error if there is a problem
func DecodeUUIDSlice(vals []string) (interface{}, error) {
	var uus []uuid.UUID

	for _, val := range vals {
		uuid, err := uuid.FromString(val)
		if err != nil {
			err = fmt.Errorf("error parsing uuid: %w", err)
			return nil, err
		}

		uus = append(uus, uuid)
	}

	return uus, nil
}
