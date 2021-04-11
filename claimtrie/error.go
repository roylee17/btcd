package claimtrie

import (
	"errors"
	"fmt"
)

var (
	// errInvalidHeight is returned when the height is invalid.
	errInvalidHeight = fmt.Errorf("invalid height")

	// errNotFound is returned when the Claim or Support is not found.
	errNotFound = fmt.Errorf("not found")

	// errDuplicate is returned when the Claim or Support already exists in the node.
	errDuplicate = fmt.Errorf("duplicate")

	// errInvalidID is returned when the ID does not conform to the format.
	errInvalidID = errors.New("ID must be a 20-character hexadecimal string")
)
