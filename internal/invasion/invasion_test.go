package invasion_test

import (
	"testing"

	"github.com/padurean/mad-aliens/internal/invasion"
	"github.com/padurean/mad-aliens/internal/world"
	"github.com/stretchr/testify/require"
)

func TestInvasion(t *testing.T) {
	// invalid argument values
	_, err := invasion.New(nil, 0, nil)
	require.EqualError(t, err, "invalid args: world must not be empty, "+
		"numberOfAliens must be greater than zero, onEvent callback must not be nil")

	// happy path
	var w world.World
	require.NoError(t, w.Read("../../world.txt"))

	var events []string
	onEvent := func(event string) {
		events = append(events, event)
	}

	inv, err := invasion.New(w, 5, onEvent)
	require.NoError(t, err)
	completionMsg := inv.Run()
	require.NotEmpty(t, completionMsg)

	require.NoError(t, inv.World.Write("../../world_after_invasion.txt"))
}
