package world

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// World holds cities per their names.
type World map[string]*City

// String returns the textual representation of a world (implements the Stringer
// interface).
func (w *World) String() string {
	if w == nil {
		return "<nil>"
	}

	var sb strings.Builder
	sb.WriteString("🌐 World:\n-----------\n")

	if len(*w) == 0 {
		sb.WriteString("All cities have been destroyed! 😱\n===========")
		return sb.String()
	}

	i := 0
	cityIcons := []string{"🌆", "🏙 ", "🌇", "🌃", "🌁", "🌉"}
	for _, city := range *w {
		fmt.Fprintf(
			&sb, "%s %s %v\n",
			cityIcons[i%len(cityIcons)], city, city.Aliens)
		i++
	}

	sb.WriteString("===========")
	return sb.String()
}

// FindGhostCities returns any city names which appear in neighbor list(s) of
// this world's cities, but which don't actually exist in this world.
// This can happen, for example, if this world is inconsistently defined.
func (w *World) FindGhostCities() map[string]struct{} {
	if w == nil {
		return nil
	}

	ghostCities := make(map[string]struct{})
	for _, city := range *w {
		for neighbor := range city.Neighbors {
			if _, ok := (*w)[neighbor]; !ok {
				ghostCities[neighbor] = struct{}{}
			}
		}
	}

	return ghostCities
}

// Read reads and populates the world from the specified file.
func (w *World) Read(filePath string) error {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("failed to open input world file '%s': %v", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(fmt.Errorf("failed to close input file %s: %w", filePath, err))
		}
	}()

	*w = make(map[string]*City)

	fileScanner := bufio.NewScanner(file)
	var lineNumber int
	for fileScanner.Scan() {
		lineNumber++
		line := fileScanner.Text()
		var city City
		if err := city.Parse(line); err != nil {
			return fmt.Errorf(
				"failed to parse city from line %d '%s': %v", lineNumber, line, err)
		}
		city.OriginalLineNumber = lineNumber
		(*w)[city.Name] = &city
	}

	if err := fileScanner.Err(); err != nil {
		return fmt.Errorf("failed to scan world file '%s': %v", filePath, err)
	}

	if len(*w) == 0 {
		return fmt.Errorf("no city has been found in fil '%s'", filePath)
	}

	return nil
}

// Write writes the world to the specified file.
func (w *World) Write(filePath string) error {
	file, err := os.OpenFile( //nolint:gosec
		filepath.Clean(filePath), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf(
			"failed to open output world file '%s': %v", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(fmt.Errorf("failed to close output file %s: %w", filePath, err))
		}
	}()

	orderedCities := make([]*City, 0, len(*w))
	for _, city := range *w {
		orderedCities = append(orderedCities, city)
	}
	sort.Slice(orderedCities, func(i, j int) bool {
		return orderedCities[i].OriginalLineNumber < orderedCities[j].OriginalLineNumber
	})

	filewriter := bufio.NewWriter(file)
	for _, city := range orderedCities {
		_, err = filewriter.WriteString(fmt.Sprintf("%s\n", city))
		if err != nil {
			return fmt.Errorf("failed to write city '%s' to file: %v", city, err)
		}
	}

	if err := filewriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush file writer: %v", err)
	}

	return nil
}
