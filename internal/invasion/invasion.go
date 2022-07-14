package invasion

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/padurean/mad-aliens/internal/world"
)

// State holds the state of an invasion.
type State struct {
	DestroyedCities         map[string]struct{}
	DeadAliens              map[int]struct{}
	TravelCountersPerAliens map[int]int
	ExhaustedAliens         map[int]struct{}
	TrappedAliensPerCity    map[string]int
}

// Invasion holds an invasion instance.
type Invasion struct {
	NumberOfAliens int

	World world.World
	State *State

	onEvent func(string)
}

// New creates a new invasion without starting it.
func New(w world.World, numberOfAliens int, onEvent func(string)) *Invasion {
	return &Invasion{
		NumberOfAliens: numberOfAliens,
		World:          w,
		State: &State{
			DestroyedCities:         make(map[string]struct{}),
			DeadAliens:              make(map[int]struct{}),
			TravelCountersPerAliens: make(map[int]int),
			ExhaustedAliens:         make(map[int]struct{}),
			TrappedAliensPerCity:    make(map[string]int),
		},
		onEvent: onEvent,
	}
}

// Run starts and runs the invasion.
func (invasion *Invasion) Run() {
	invasion.landAliens()
	invasion.assess()
	if invasion.isComplete() {
		worldBytes, _ := json.MarshalIndent(invasion, "", "  ")
		invasion.onEvent(fmt.Sprintf(
			"Invasion complete right after aliens landing! "+
				"State of the world:\n%s", worldBytes))
		return
	}

	for {
		invasion.advance()
		invasion.assess()
		if invasion.isComplete() {
			worldBytes, _ := json.MarshalIndent(invasion, "", "  ")
			invasion.onEvent(fmt.Sprintf(
				"Invasion complete! State of the world:\n%s", worldBytes))
			return
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

	worldBytes, _ := json.MarshalIndent(invasion, "", "  ")
	invasion.onEvent(fmt.Sprintf(
		"Aliens landed! State of the world:\n%s", worldBytes))
}

func (invasion *Invasion) advance() {
	for _, city := range invasion.World {
		if len(city.Aliens) == 0 || len(city.Neighbors) == 0 {
			continue
		}

		rand.Seed(time.Now().UnixNano())
		var nextCity *world.City
		nextCityIndex := rand.Intn(len(city.Neighbors))
		i := 0
		for neighborName := range city.Neighbors {
			if i == nextCityIndex {
				nextCity = invasion.World[neighborName]
				break
			}
			i++
		}

		// TODO OGG NOW: assessment should happen right here, otherwise aliens can pass
		// through cities without fighting.
		// Find a solution to do this assessment also right after landing (e.g. if 2
		// or more aliens land in the same city)
		travelingAlien := city.Aliens[0]
		nextCity.Aliens = append(nextCity.Aliens, travelingAlien)
		city.Aliens = city.Aliens[1:]
		invasion.State.TravelCountersPerAliens[travelingAlien]++

		invasion.onEvent(fmt.Sprintf(
			"Alien %d has traveled from %+v (aliens: %v) to %+v (aliens: %v)",
			travelingAlien, city, city.Aliens, nextCity, nextCity.Aliens))
	}
}

// assess assesses destroyed cities, dead and trapped aliens and updates the
// state of the invasion and of the world.
func (invasion *Invasion) assess() {
	invasion.collectDestroyedCitiesAndDeadAliens()
	invasion.collectExhaustedAliens()

}

func (invasion *Invasion) collectDestroyedCitiesAndDeadAliens() {
	destroyedCities := make(map[string]struct{})

	for _, city := range invasion.World {
		if len(city.Aliens) > 1 {
			destroyedCities[city.Name] = struct{}{}
			invasion.State.DestroyedCities[city.Name] = struct{}{}
			aliensNames := make([]string, 0, len(city.Aliens))
			for _, alien := range city.Aliens {
				invasion.State.DeadAliens[alien] = struct{}{}
				aliensNames = append(aliensNames, fmt.Sprintf("alien %d", alien))
			}
			invasion.emitCityDestructionEvent(city.Name, aliensNames)
		}
	}

	invasion.removeCities(destroyedCities)
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

func (invasion *Invasion) removeCities(cities map[string]struct{}) {
	for name := range cities {
		delete(invasion.World, name)
	}

	for _, city := range invasion.World {
		for neighbor := range city.Neighbors {
			if _, ok := cities[neighbor]; ok {
				delete(city.Neighbors, neighbor)
			}
		}
		if len(city.Neighbors) == 0 && len(city.Aliens) > 0 {
			invasion.State.TrappedAliensPerCity[city.Name] = city.Aliens[0]
		}
	}
}

func (invasion *Invasion) collectExhaustedAliens() {
	for alien, travelCounter := range invasion.State.TravelCountersPerAliens {
		if travelCounter == 10000 {
			invasion.State.ExhaustedAliens[alien] = struct{}{}
		}
	}
}

func (invasion *Invasion) isComplete() bool {
	if len(invasion.State.DestroyedCities) == len(invasion.World) {
		return true
	}
	if len(invasion.State.DeadAliens)+
		len(invasion.State.ExhaustedAliens)+
		len(invasion.State.TrappedAliensPerCity)+1 >= invasion.NumberOfAliens {
		return true
	}
	return false
}
