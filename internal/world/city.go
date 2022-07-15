package world

import (
	"fmt"
	"strings"
)

// City holds a city name and it's neighboring cities names per directions.
type City struct {
	Name               string
	Neighbors          map[string]Direction
	OriginalLineNumber int
	Aliens             []int
}

// String returns the textual representation of a city (implements the Stringer
// interface).
func (c *City) String() string {
	if c == nil {
		return "<nil>"
	}

	var sb strings.Builder
	sb.WriteString(c.Name)
	for name, direction := range c.Neighbors {
		sb.WriteString(" ")
		sb.WriteString(direction.String())
		sb.WriteString("=")
		sb.WriteString(name)
	}

	return sb.String()
}

// Parse populates the city from the specified textual representation.
func (c *City) Parse(s string) error {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return fmt.Errorf("failed to parse city: empty string")
	}

	c.Neighbors = make(map[string]Direction)

	for i, w := range strings.Split(s, " ") {
		if i == 0 {
			c.Name = w
			continue
		}

		directionAndName := strings.SplitN(w, "=", 2)
		if len(directionAndName) < 2 {
			return fmt.Errorf(
				"failed to parse direction and neighbor city from '%s': "+
					"expected format <direction>=<city-name>", w)
		}

		var d Direction
		if !d.Parse(directionAndName[0]) {
			return fmt.Errorf(
				"failed to parse direction from string '%s'", directionAndName[0])
		}

		c.Neighbors[directionAndName[1]] = d
	}

	return nil
}

// Removes any neighbors that are present among the specified cities.
func (c *City) RemoveNeighborsIn(cities map[string]struct{}) {
	if len(cities) == 0 {
		return
	}

	for neighbor := range c.Neighbors {
		if _, ok := cities[neighbor]; ok {
			delete(c.Neighbors, neighbor)
		}
	}
}
