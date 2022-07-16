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
	DeadAliens              map[string]struct{}
	TravelCountersPerAliens map[string]int
	ExhaustedAliens         map[string]struct{}
	TrappedAliensPerCity    map[string][]string
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
	fmt.Fprintf(&sb, "💥 Destroyed cities:           %v\n", s.DestroyedCities)
	fmt.Fprintf(&sb, "☠️  Dead aliens:                %v\n", s.DeadAliens)
	fmt.Fprintf(&sb, "⚡️ Travel counters per aliens: %v\n", s.TravelCountersPerAliens)
	fmt.Fprintf(&sb, "🚷 Trapped aliens per city:    %v\n", s.TrappedAliensPerCity)
	fmt.Fprintf(&sb, "😪 Exhausted aliens:           %v\n", s.ExhaustedAliens)
	fmt.Fprintf(&sb, "👻 Ghost cities:               %v\n", s.GhostCities)

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
func New(
	w world.World,
	numberOfAliens int,
	onEvent func(string),
) (*Invasion, error) {

	var validationErrs []string
	if len(w) == 0 {
		validationErrs = append(validationErrs, "world must not be empty")
	}
	if numberOfAliens <= 0 {
		validationErrs = append(
			validationErrs, "numberOfAliens must be greater than zero")
	}
	if onEvent == nil {
		validationErrs = append(validationErrs, "onEvent callback must not be nil")
	}
	if len(validationErrs) > 0 {
		return nil, fmt.Errorf("invalid args: %s", strings.Join(validationErrs, ", "))
	}

	invasion := &Invasion{
		NumberOfAliens: numberOfAliens,
		World:          w,
		Summary: &Summary{
			DestroyedCities:         make(map[string]struct{}),
			DeadAliens:              make(map[string]struct{}),
			TravelCountersPerAliens: make(map[string]int),
			ExhaustedAliens:         make(map[string]struct{}),
			TrappedAliensPerCity:    make(map[string][]string),
			GhostCities:             w.FindGhostCities(),
		},
		onEvent: onEvent,
	}

	// Set real neighbors for each city.
	if len(invasion.Summary.GhostCities) > 0 {
		for _, city := range invasion.World {
			city.SetRealNeighborsFromGhosts(invasion.Summary.GhostCities)
		}
	}

	return invasion, nil
}

// Run starts and runs the invasion.
func (invasion *Invasion) Run() string {
	invasion.landAliens()
	if invasion.advance(false) {
		return fmt.Sprintf(
			"☑️  Invasion complete right after aliens landing!\n%s",
			invasion.String())
	}

	for {
		if invasion.advance(true) {
			return fmt.Sprintf("☑️  Invasion complete!\n%s", invasion.String())
		}
	}
}

func (invasion *Invasion) landAliens() {
	alienTeams := make([][]string, len(invasion.World))

	var alienCounter int
	for {
		rand.Seed(time.Now().UnixNano())
		nextTeamIndex := rand.Intn(len(alienTeams)) //nolint:gosec
		alienCounter++
		alienTeams[nextTeamIndex] =
			append(alienTeams[nextTeamIndex], fmt.Sprintf("👽%d", alienCounter))
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
		"🛸 %d 👽 aliens landed!\n%s", invasion.NumberOfAliens, invasion.World.String()))
}

func (invasion *Invasion) advance(teleportAliens bool) bool {
	var invasionIsComplete bool

	for _, city := range invasion.World {
		if len(city.Aliens) == 0 {
			continue
		}

		// Collect the destroyed city and it's dead aliens.
		if len(city.Aliens) > 1 {
			invasion.destroyCity(city)
			continue
		}

		if len(city.RealNeighbors) == 0 || !teleportAliens {
			continue
		}

		// Teleport alien to a next random city.
		nextCityName, _ := city.GetRandomNeighbor()
		nextCity := invasion.World[nextCityName]
		invasion.teleportAlien(city, nextCity)

		// Collect the next destroyed city and it's dead aliens.
		if len(nextCity.Aliens) > 1 {
			invasion.destroyCity(nextCity)
		}

		invasionIsComplete = invasion.isComplete()
		if invasionIsComplete {
			break
		}
	}

	return invasionIsComplete || invasion.isComplete()
}

// teleportAlien "teleports" an alien from once city to another.
// It also updates the state by incrementing the alien's travel counters and by
// marking the alien as exhausted (if it reached the exhaustion limit).
// At the end it emits a travel event.
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
		"⚡️ %s traveled   from  ➡️  %+v %v   to  ➡️  %+v %v",
		alien, from, from.Aliens, to, to.Aliens))
}

// destroyCity collects a destroyed city and it's dead aliens.
// It also removes the city from other cities neighbors lists and it emits a
// city destruction event at the end.
// !NOTE: The caller must check that the city has indeed been destroyed.
func (invasion *Invasion) destroyCity(city *world.City) {
	delete(invasion.World, city.Name)
	delete(invasion.Summary.TrappedAliensPerCity, city.Name)
	invasion.Summary.DestroyedCities[city.Name] = struct{}{}

	// Remove the city from other cities neighbors and update the trapped aliens.
	for _, otherCity := range invasion.World {
		if len(otherCity.RealNeighbors) == 0 {
			continue
		}
		otherCity.RemoveNeighbor(city.Name)
		if len(otherCity.RealNeighbors) == 0 && len(otherCity.Aliens) > 0 {
			invasion.Summary.TrappedAliensPerCity[otherCity.Name] = append(
				invasion.Summary.TrappedAliensPerCity[otherCity.Name], otherCity.Aliens...)
		}
	}
	// !NOTE: if the world map is guaranteed to be defined in a consistent manner
	// (meaning that for each and every neighboring cities X and Y
	// X appears in the list of Y's neighbors AND vice versa), then the above
	// removal of neighbor could be made quicker by iterating only over the
	// destroyed city's neighbors instead of iterating over all the cities,
	// as follows:
	// for neighbor := range city.RealNeighbors {
	// 	invasion.World[neighbor].RemoveNeighbor(city.Name)
	// }

	aliensNames := make([]string, 0, len(city.Aliens))
	for _, alien := range city.Aliens {
		invasion.Summary.DeadAliens[alien] = struct{}{}
		aliensNames = append(aliensNames, alien)
	}
	invasion.emitDestructionEvent(city.Name, aliensNames)
}

func (invasion *Invasion) emitDestructionEvent(city string, aliens []string) {
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

func (invasion *Invasion) isComplete() bool {
	if len(invasion.Summary.DestroyedCities) == len(invasion.World) {
		return true
	}

	var nbTrappedAliens int
	for _, aliens := range invasion.Summary.TrappedAliensPerCity {
		nbTrappedAliens += len(aliens)
	}

	return len(invasion.Summary.DeadAliens)+
		len(invasion.Summary.ExhaustedAliens)+
		nbTrappedAliens+1 >= invasion.NumberOfAliens
}
