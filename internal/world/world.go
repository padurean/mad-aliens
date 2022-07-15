package world

import (
	"bufio"
	"fmt"
	"os"
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
	sb.WriteString("World:\n---\n")

	if len(*w) == 0 {
		sb.WriteString("All cities have been destroyed! 😱\n---")
		return sb.String()
	}

	i := 0
	for _, city := range *w {
		fmt.Fprintf(&sb, "%s (aliens: %v)", city, city.Aliens)
		if i < len(*w)-1 {
			sb.WriteString("\n")
		}
		i++
	}

	sb.WriteString("\n---")
	return sb.String()
}

// Read reads and populates the world from the specified file.
func (w *World) Read(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open input world file '%s': %v", filePath, err)
	}
	defer file.Close()

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
	file, err := os.OpenFile(filePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf(
			"failed to open output world file '%s': %v", filePath, err)
	}
	defer file.Close()

	orderedCities := make([]*City, 0, len(*w))
	for _, city := range *w {
		orderedCities = append(orderedCities, city)
	}
	sort.Slice(orderedCities, func(i, j int) bool {
		return orderedCities[i].OriginalLineNumber < orderedCities[j].OriginalLineNumber
	})

	datawriter := bufio.NewWriter(file)
	for _, city := range orderedCities {
		_, err = datawriter.WriteString(fmt.Sprintf("%s\n", city))
		if err != nil {
			return fmt.Errorf("failed to write city '%s' to file: %v", city, err)
		}
	}
	datawriter.Flush()

	return nil
}
