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

:exclamation: The world is read from the [world.txt](./world.txt) file and what's left of it after the invasion simulation is written to [world_after_invasion.txt](./world_after_invasion.txt).

:bulb: A [Makefile](./Makefile) is also provided for convenience. Please have a look at it for all the available targets, among which:

- ***make run_dev***
- ***make lint*** (:exclamation: requires [golangci-lint](https://golangci-lint.run) to be installed)
- ***make test***
- ***make coverage-report***

:bulb: A [Dockerfile](./Dockerfile) is provided for running it with [Docker](https://www.docker.com) and some convenience make targets like ***make build_docker_image***. To run it with Docker: `docker run padurean/mad-aliens 10`
