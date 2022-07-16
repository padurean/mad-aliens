package world

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// City holds a city name and it's neighboring cities names per directions.
type City struct {
	Name               string
	Neighbors          map[string]Direction
	OriginalLineNumber int
	Aliens             []string

	// Some neighbors might be "ghosts" - i.e. they might not actually exist.
	// If that is the case, this field can be populated with just the real
	// neighbors.
	RealNeighbors map[string]Direction
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

// SetRealNeighborsFromGhosts populates this city's RealNeighbors fields with
// only the neighbors that do not occur in the specified ghost cities list.
func (c *City) SetRealNeighborsFromGhosts(ghostCities map[string]struct{}) {
	c.RealNeighbors = make(map[string]Direction)
	for name, direction := range c.Neighbors {
		if _, ok := ghostCities[name]; !ok {
			c.RealNeighbors[name] = direction
		}
	}
}

// GetRandomNeighbor returns a random (real) neighbor.
// !NOTE: The caller must ensure that city has at least one neighbor.
func (c *City) GetRandomNeighbor() (string, Direction) {
	neighbors := c.RealNeighbors
	if len(neighbors) == 0 {
		neighbors = c.Neighbors
	}
	if len(neighbors) == 0 {
		return "", DirectionUnknown
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(neighbors))

	i := 0
	for neighbor, direction := range neighbors {
		if i == randomIndex {
			return neighbor, direction
		}
		i++
	}

	return "", DirectionUnknown
}

// RemoveNeighbor removes a neighbor city from this city neighbor lists.
func (c *City) RemoveNeighbor(neighbor string) {
	delete(c.Neighbors, neighbor)
	delete(c.RealNeighbors, neighbor)
}
