package invasion

import (
	"fmt"
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
	GhostCities             map[string]struct{}
}

// String returns the textual representation of a summary (implements the
// Stringer interface).
func (s *Summary) String() string {
	if s == nil {
		return "<nil>"
	}

	var sb strings.Builder
	sb.WriteString("📜 Summary:\n-----------\n")
	fmt.Fprintf(&sb, "💥 Destroyed cities          : %v\n", s.DestroyedCities)
	fmt.Fprintf(&sb, "☠️  Dead aliens               : %v\n", s.DeadAliens)
	fmt.Fprintf(&sb, "🔂 Travel counters per aliens: %v\n", s.TravelCountersPerAliens)
	fmt.Fprintf(&sb, "😪 Exhausted aliens          : %v\n", s.ExhaustedAliens)
	fmt.Fprintf(&sb, "🚷 Trapped aliens per city   : %v\n", s.TrappedAliensPerCity)
	fmt.Fprintf(&sb, "👻 Ghost cities              : %v\n", s.GhostCities)

	sb.WriteString("===========")
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
	invasion := &Invasion{
		NumberOfAliens: numberOfAliens,
		World:          w,
		Summary: &Summary{
			DestroyedCities:         make(map[string]struct{}),
			DeadAliens:              make(map[int]struct{}),
			TravelCountersPerAliens: make(map[int]int),
			ExhaustedAliens:         make(map[int]struct{}),
			TrappedAliensPerCity:    make(map[string][]int),
			GhostCities:             w.FindGhostCities(),
		},
		onEvent: onEvent,
	}

	// set real neighbors for each city
	if len(invasion.Summary.GhostCities) > 0 {
		for _, city := range invasion.World {
			city.SetRealNeighborsFromGhosts(invasion.Summary.GhostCities)
		}
	}

	return invasion
}

// Run starts and runs the invasion.
func (invasion *Invasion) Run() string {
	invasion.landAliens()
	if invasion.advance(false) {
		return fmt.Sprintf(
			"Invasion complete right after aliens landing!\n%s",
			invasion.String())
	}

	for {
		if invasion.advance(true) {
			return fmt.Sprintf("Invasion complete!\n%s", invasion.String())
		}
	}
}

func (invasion *Invasion) landAliens() {
	alienTeams := make([][]int, len(invasion.World))

	var alienCounter int
	for {
		rand.Seed(time.Now().UnixNano())
		nextTeamIndex := rand.Intn(len(alienTeams))
		alienCounter++
		alienTeams[nextTeamIndex] = append(alienTeams[nextTeamIndex], alienCounter)
		if alienCounter == invasion.NumberOfAliens {
			break
		}
	}

	i := 0
	for _, city := range invasion.World {
		city.Aliens = alienTeams[i]
		i++
	}

	invasion.onEvent(fmt.Sprintf(
		"%d 👽 aliens landed!\n%s", invasion.NumberOfAliens, invasion.World.String()))
}

func (invasion *Invasion) advance(teleportAliens bool) bool {
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
		if len(city.RealNeighbors) == 0 || !teleportAliens {
			continue
		}

		nextCityName, _ := city.GetRandomNeighbor()
		nextCity := invasion.World[nextCityName]
		// This should never happen.
		if nextCity == nil {
			continue
		}
		invasion.teleportAlien(city, nextCity)

		// Collect the next destroyed city and it's dead aliens.
		if len(nextCity.Aliens) > 1 {
			invasion.destroyCity(nextCity)
		}

		if invasion.isComplete() {
			invasion.removeDestroyedNeighbors()
			return true
		}
	}

	invasion.removeDestroyedNeighbors()
	return invasion.isComplete()
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
		"Alien %d has been teleported from { %+v 👽%v } ➡️ to { %+v 👽%v }",
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
		aliensNames = append(aliensNames, fmt.Sprintf("👽%d", alien))
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
		"💥 %s has been destroyed by %s!", city, aliensJoined))
}

// removeDestroyedNeighbors removes any destroyed cities from the neighbor(s)
// lists of other cities.
// It also updates the state by collecting the trapped aliens (for all cities which
// are left without neighbors.
func (invasion *Invasion) removeDestroyedNeighbors() {
	for _, city := range invasion.World {
		city.RemoveNeighborsIn(invasion.Summary.DestroyedCities)
		if len(city.RealNeighbors) == 0 && len(city.Aliens) > 0 {
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
