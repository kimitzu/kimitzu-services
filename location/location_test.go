package location

import (
	"strconv"
	"testing"
)

func TestDistanceCalculationLocal(t *testing.T) {
	// PH, Iloilo City
	sourceX := 10.6969
	sourceY := 122.5644
	// PH, Pavia
	targetX := 10.7761
	targetY := 122.5456

	expectedDistace, _ := strconv.ParseFloat("9053.04491084398", 64)
	distance := Distance(sourceX, sourceY, targetX, targetY)

	if distance != expectedDistace {
		t.Errorf("Distance from Location(%.4f, %.4f) to Location(%.4f, %.4f) is %.4f, expected %.4f", sourceX, sourceY, targetX, targetY, distance, expectedDistace)
	}
}

func TestDistanceCalculationInternational(t *testing.T) {
	// PH, Iloilo City
	sourceX := 10.6969
	sourceY := 122.5644
	// AU, Adelaide River Northern Territory NT DARWIN
	targetX := -13.2379
	targetY := 131.1056

	expectedDistace, _ := strconv.ParseFloat("2826546.0938833216", 64)
	distance := Distance(sourceX, sourceY, targetX, targetY)

	if distance != expectedDistace {
		t.Errorf("Distance from Location(%.4f, %.4f) to Location(%.4f, %.4f) is %.4f, expected %.4f", sourceX, sourceY, targetX, targetY, distance, expectedDistace)
	}
}
