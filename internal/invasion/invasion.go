package invasion

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/padurean/mad-aliens/internal/world"
)

const exhaustionLimit = 10000

// Summary holds the summary of an invasion.
type Summary struct {
	DestroyedCities         map[string]struct{}
	DeadAliens              map[int]struct{}
	TravelCountersPerAliens map[int]int
	ExhaustedAliens         map[int]struct{}
	TrappedAliensPerCity    map[string][]int
}

// String returns the textual representation of a summary (implements the
// Stringer interface).
func (s *Summary) String() string {
	if s == nil {
		return "<nil>"
	}

	var sb strings.Builder
	sb.WriteString("Summary:\n---\n")
	fmt.Fprintf(&sb, "Destroyed cities          : %v\n", s.DestroyedCities)
	fmt.Fprintf(&sb, "Dead aliens               : %v\n", s.DeadAliens)
	fmt.Fprintf(&sb, "Travel counters per aliens: %v\n", s.TravelCountersPerAliens)
	fmt.Fprintf(&sb, "Exhausted aliens          : %v\n", s.ExhaustedAliens)
	fmt.Fprintf(&sb, "Trapped aliens per city   : %v", s.TrappedAliensPerCity)

	sb.WriteString("\n---")
	return sb.String()
}

// Invasion holds an invasion instance.
type Invasion struct {
	NumberOfAliens int

	World   world.World
	Summary *Summary

	onEvent func(string)
}

// String returns the textual representation of an invasion (implements the
// Stringer interface).
func (invasion *Invasion) String() string {
	if invasion == nil {
		return "<nil>"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", invasion.World.String())
	fmt.Fprintf(&sb, "%s", invasion.Summary.String())

	return sb.String()
}

// New creates a new invasion without starting it.
func New(w world.World, numberOfAliens int, onEvent func(string)) *Invasion {
	return &Invasion{
		NumberOfAliens: numberOfAliens,
		World:          w,
		Summary: &Summary{
			DestroyedCities:         make(map[string]struct{}),
			DeadAliens:              make(map[int]struct{}),
			TravelCountersPerAliens: make(map[int]int),
			ExhaustedAliens:         make(map[int]struct{}),
			TrappedAliensPerCity:    make(map[string][]int),
		},
		onEvent: onEvent,
	}
}

// Run starts and runs the invasion.
func (invasion *Invasion) Run() string {
	invasion.landAliens()
	invasion.advance(false)
	if invasion.isComplete() {
		return fmt.Sprintf(
			"Invasion complete right after aliens landing!\n%s",
			invasion.String())
	}

	for {
		invasion.advance(true)
		if invasion.isComplete() {
			return fmt.Sprintf("Invasion complete!\n%s", invasion.String())
		}
	}
}

func (invasion *Invasion) landAliens() {
	numberOfLoops :=
		int(math.Ceil(float64(invasion.NumberOfAliens) / float64(len(invasion.World))))

	var alien int
	for i := 0; i < numberOfLoops; i++ {
		for _, city := range invasion.World {
			alien++
			city.Aliens = append(city.Aliens, alien)
			if alien == invasion.NumberOfAliens {
				break
			}
		}
	}

	invasion.onEvent(fmt.Sprintf(
		"%d aliens👽 landed!\n%s", invasion.NumberOfAliens, invasion.World.String()))
}

func (invasion *Invasion) advance(teleportAliens bool) {
	for _, city := range invasion.World {
		if len(city.Aliens) == 0 {
			continue
		}

		// Collect the destroyed city and it's dead aliens.
		if len(city.Aliens) > 1 {
			invasion.destroyCity(city)
			continue
		}

		// Remove any meanwhile destroyed cities from the list of neighbors.
		city.RemoveNeighborsIn(invasion.Summary.DestroyedCities)
		if len(city.Neighbors) == 0 || !teleportAliens {
			continue
		}

		nextCity := invasion.getRandomNeighboringCity(city)
		// Next city can be nil if the input map is inconsistent - i.e. if it has
		// neighbor(s) list(s) that contain non-existent cities.
		if nextCity == nil {
			continue
		}
		invasion.teleportAlien(city, nextCity)

		// Collect the next destroyed city and it's dead aliens.
		if len(nextCity.Aliens) > 1 {
			invasion.destroyCity(nextCity)
		}
	}

	invasion.removeDestroyedNeighbors()
}

// getRandomNeighboringCity picks and returns a random neighbor of the
// specified city.
// !NOTE: The caller must ensure that city has at least one neighbor.
func (invasion *Invasion) getRandomNeighboringCity(city *world.City) *world.City {
	rand.Seed(time.Now().UnixNano())
	nextCityIndex := rand.Intn(len(city.Neighbors))

	var nextCity *world.City

	i := 0
	for neighborName := range city.Neighbors {
		if i == nextCityIndex {
			nextCity = invasion.World[neighborName]
			break
		}
		i++
	}

	return nextCity
}

// teleportAlien "teleports" an alien from once city to another.
// It also updates the state by incrementing the alien's travel counters and by
// marking the alien as exhausted (if it reached the exhaustion limit).
// At the end it emits an alien "has been teleported" event.
func (invasion *Invasion) teleportAlien(from, to *world.City) {
	alien := from.Aliens[0]
	// Exhausted aliens can't travel anymore.
	if _, ok := invasion.Summary.ExhaustedAliens[alien]; ok {
		return
	}

	to.Aliens = append(to.Aliens, alien)
	from.Aliens = from.Aliens[1:]

	invasion.Summary.TravelCountersPerAliens[alien]++
	if invasion.Summary.TravelCountersPerAliens[alien] == exhaustionLimit {
		invasion.Summary.ExhaustedAliens[alien] = struct{}{}
	}

	invasion.onEvent(fmt.Sprintf(
		"Alien %d has been teleported from %+v (aliens: %v) to %+v (aliens: %v)",
		alien, from, from.Aliens, to, to.Aliens))
}

// destroyCity collects a destroyed city and it's dead aliens.
// Also emits a city destruction event.
// !NOTE: The caller must check that the city has indeed been destroyed.
func (invasion *Invasion) destroyCity(city *world.City) {
	delete(invasion.World, city.Name)
	invasion.Summary.DestroyedCities[city.Name] = struct{}{}

	aliensNames := make([]string, 0, len(city.Aliens))
	for _, alien := range city.Aliens {
		invasion.Summary.DeadAliens[alien] = struct{}{}
		aliensNames = append(aliensNames, fmt.Sprintf("alien %d", alien))
	}
	invasion.emitCityDestructionEvent(city.Name, aliensNames)
}

func (invasion *Invasion) emitCityDestructionEvent(city string, aliens []string) {
	var aliensJoined string
	if len(aliens) == 2 {
		aliensJoined = strings.Join(aliens, " and ")
	} else {
		aliensJoined = strings.Join(aliens[:len(aliens)-1], ", ")
		aliensJoined += " and " + aliens[len(aliens)-1]
	}
	invasion.onEvent(fmt.Sprintf(
		"%s has been destroyed by %s!", city, aliensJoined))
}

// removeDestroyedNeighbors removes any destroyed cities from the neighbor(s)
// lists of other cities.
// It also updates the state by collecting the trapped aliens (for all cities which
// are left without neighbors.
func (invasion *Invasion) removeDestroyedNeighbors() {
	for _, city := range invasion.World {
		city.RemoveNeighborsIn(invasion.Summary.DestroyedCities)
		if len(city.Neighbors) == 0 && len(city.Aliens) > 0 {
			invasion.Summary.TrappedAliensPerCity[city.Name] = append(
				invasion.Summary.TrappedAliensPerCity[city.Name], city.Aliens...)
		}
	}
}

func (invasion *Invasion) isComplete() bool {
	if len(invasion.Summary.DestroyedCities) == len(invasion.World) {
		return true
	}

	var nbTrappedAliens int
	for _, aliens := range invasion.Summary.TrappedAliensPerCity {
		nbTrappedAliens += len(aliens)
	}

	if len(invasion.Summary.DeadAliens)+
		len(invasion.Summary.ExhaustedAliens)+
		nbTrappedAliens+1 >= invasion.NumberOfAliens {
		return true
	}
	return false
}
