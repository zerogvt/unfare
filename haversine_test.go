package main

import (
	"math"
	"testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestDegreesToRadians(t *testing.T) {
	have := int(math.Round(100 * degreesToRadians(180)))
	wantx100 := 314
	if wantx100 != have {
		t.Fatalf("Wanted %d but got %d", wantx100, have)
	}
}

func TestDistance(t *testing.T) {
	want := 14
	have := int(Distance(Coordinates{37.94, 23.63}, Coordinates{37.94, 23.80}))
	if want != have {
		t.Fatalf("Wanted %d but got %d", want, have)
	}
}
