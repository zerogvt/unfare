package main

import (
	"math"
	"testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestVelocity(t *testing.T) {
	p1 := Point{1, Coordinates{37.9360466003418, 23.94614028930664}, 1405090921}
	p2 := Point{1, Coordinates{37.93638229370117, 23.94644546508789}, 1405090930}
	have := int(math.Round(p1.velocity(&p2)))
	want := 18
	if want != have {
		t.Fatalf("TestVelocity: Wanted %d but got %d", want, have)
	}
}

func TestVelocityZero(t *testing.T) {
	p1 := Point{1, Coordinates{37.9360466003418, 23.94614028930664}, 1405090921}
	p2 := Point{1, Coordinates{37.9360466003418, 23.94614028930664}, 1405090930}
	have := int(math.Round(p1.velocity(&p2)))
	want := 0
	if want != have {
		t.Fatalf("TestVelocityZero: Wanted %d but got %d", want, have)
	}
}

func TestFareTo(t *testing.T) {
	p1 := Point{1, Coordinates{37.910030364990234, 23.90641212463379}, 1405090726}
	p2 := Point{1, Coordinates{37.93056106567383, 23.93911361694336}, 1405090858}
	have, _ := p1.fareTo(&p2)
	havex100 := int(math.Round(100 * have))
	wantx100 := 271
	if wantx100 != havex100 {
		t.Fatalf("TestFareTo: Wanted %d but got %d", wantx100, havex100)
	}
}

func TestFareToWeedsOutlierPoint(t *testing.T) {
	p1 := Point{1, Coordinates{37.910030364990234, 23.90641212463379}, 1405090726}
	// p2 longtitude skew
	p2 := Point{1, Coordinates{37.93056106567383, 28.93911361694336}, 1405090858}
	have, err := p1.fareTo(&p2)
	if err == nil {
		t.Fatalf("TestFareTo did not weed out outlier point")
	}
	if int(have) != 0 {
		t.Fatalf("TestFareTo did not return zero fair for outlier point")
	}
}
