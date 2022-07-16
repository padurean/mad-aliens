# :alien::alien::alien: Mad Aliens

Simulating an alien invasion using the [Go]((https://go.dev)) programming language.

The requirements can be found [here](./requirements.md).

## Prerequisites

### Required

- [Go](https://go.dev)

### Optional

- [golangci-lint](https://golangci-lint.run)
- [make](https://www.cs.swarthmore.edu/~newhall/unixhelp/howto_makefiles.html)
- [Docker](https://www.docker.com)

## Run, Lint, Test and more

Run it using `go run`, passing the number of aliens:

`go run ./cmd/invasion 5`

or to run it interactively, step by step, also pass the `-i` flag:

`go run ./cmd/invasion -i 12`

:exclamation: The world is read from the [world.txt](./world.txt) file and what's left of it after the invasion simulation is written to a ***world_after_invasion.txt*** file.

:bulb: A [Makefile](./Makefile) is also provided for convenience. Please have a look at it for all the available targets, among which:

- ***make run_dev***
- ***make lint*** (:exclamation: requires [golangci-lint](https://golangci-lint.run) to be installed)
- ***make test***
- ***make coverage-report***

:bulb: A [Dockerfile](./Dockerfile) is provided for running it with [Docker](https://www.docker.com) and some convenience make targets like ***make build_docker_image***. To run it with Docker: `docker run padurean/mad-aliens 10`

## Code and package design

The code is structured as follows:

- [cmd/invasion/](./cmd/invasion/main.go) is the main package containing just the [main.go](/cmd/invasion/main.go) file. It's code parses the CLI flag and argument, reads and parse the world from the [world.txt](./world.txt) file, creates and runs a new invasion and then writes what's lef of the world to a [world_after_invasion.txt] file. When creating the invasion, it also passes an `onEvent` callback to it which prints any events (basically informative strings) to the console.

- [internal/world](./internal/world/) package contains structs and methods related to the world (map):
  - a custom [`World`](./internal/world/world.go) type definition (for a map holding the cities per their names) with convenience methods like `String`, `Read` and `Write` (from/to file) and also a more exotic one `FindGhostCities`:ghost:
  - a [`City`](./internal/world/direction.go) struct type which holds a city name, neighbors directions per names and aliens names.
  - a [`Direction`](./internal/world/direction.go) enum type for convenience and a more type-safe way for dealing with related, fixed values.

- [internal/invasion](./internal/invasion/) package contains a main [`Invasion`](./internal/invasion/invasion.go) struct representing the invasion itself with 2 exported methods (beside `String`): `New` for constructing it and `Run` to run it. It also has some non-exported methods which perform specialized tasks and update the state of the world and the summary of the invasion.

For more details please have a look over the code itself. Enjoy! :wave:
