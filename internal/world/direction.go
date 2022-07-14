package world

import (
	"fmt"
	"strings"
)

// Direction represents a geographical direction such as east, west, north or
// south.
type Direction int

// Direction values
const (
	DirectionUnknown Direction = iota
	DirectionEast
	DirectionWest
	DirectionNorth
	DirectionSouth
)

// String returns the textual representation of a direction (implements the
// Stringer interface).
func (d Direction) String() string {
	return [...]string{"unknown", "east", "west", "north", "south"}[d]
}

// Parse sets the direction value based on the specified textual representation.
func (d *Direction) Parse(str string) bool {
	dd, ok := map[string]Direction{
		"east":  DirectionEast,
		"west":  DirectionWest,
		"north": DirectionNorth,
		"south": DirectionSouth,
	}[strings.ToLower(str)]

	if ok {
		*d = dd
	}

	return ok
}

// UnmarshalText implements the standard library interface
// encoding.TextUnmarshaler.
func (d *Direction) UnmarshalText(text []byte) error {
	if d.Parse(string(text)) {
		return nil
	}
	return fmt.Errorf("unknown direction '%s'", text)
}

// MarshalText implements the standard library interface encoding.TextMarshaler.
func (d Direction) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}
