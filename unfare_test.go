package main

import (
	"io/ioutil"
	"math"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestVelocity(t *testing.T) {
	p1 := Point{1, Coordinates{37.93604, 23.94614}, 1405090921}
	p2 := Point{1, Coordinates{37.93638, 23.94644}, 1405090930}
	have := int(math.Round(p1.velocity(&p2)))
	want := 18
	if want != have {
		t.Fatalf("Wanted %d but got %d", want, have)
	}
}

func TestVelocityZero(t *testing.T) {
	p1 := Point{1, Coordinates{37.93604, 23.94614}, 1405090921}
	p2 := Point{1, Coordinates{37.93604, 23.94614}, 1405090930}
	have := int(math.Round(p1.velocity(&p2)))
	want := 0
	if want != have {
		t.Fatalf("Wanted %d but got %d", want, have)
	}
}

func TestFareToDaily(t *testing.T) {
	p1 := Point{1, Coordinates{37.91003, 23.90641}, 1405090726}
	p2 := Point{1, Coordinates{37.93056, 23.93911}, 1405090858}
	have, _ := p1.fareTo(&p2)
	havex100 := int(math.Round(100 * have))
	wantx100 := 271
	// we don't want to compare floats
	if wantx100 != havex100 {
		t.Fatalf("Wanted %d but got %d", wantx100, havex100)
	}
}

func TestFareToNightly(t *testing.T) {
	p1 := Point{1, Coordinates{37.91003, 23.90641}, 1636158734}
	p2 := Point{1, Coordinates{37.930561, 23.93911}, 1636159814}
	have, _ := p1.fareTo(&p2)
	havex100 := int(math.Round(100 * have))
	wantx100 := 477
	if wantx100 != havex100 {
		t.Fatalf("Wanted %d but got %d", wantx100, havex100)
	}
}

func TestFareToIdling(t *testing.T) {
	p1 := Point{1, Coordinates{37.91003, 23.90641}, 1636158734}
	p2 := Point{1, Coordinates{37.93056, 23.90641}, 1636159814}
	have, _ := p1.fareTo(&p2)
	havex100 := int(math.Round(100 * have))
	wantx100 := 357
	if wantx100 != havex100 {
		t.Fatalf("Wanted %d but got %d", wantx100, havex100)
	}
}

func TestFareToWeedsOutlierPoint(t *testing.T) {
	p1 := Point{1, Coordinates{37.91003, 23.90641}, 1405090726}
	// p2 longtitude skew
	p2 := Point{1, Coordinates{37.93056, 28.93911}, 1405090858}
	have, err := p1.fareTo(&p2)
	if err == nil {
		t.Fatalf("Did not weed out outlier point")
	}
	if int(have) != 0 {
		t.Fatalf("Did not return zero fair for outlier point")
	}
}

func TestNewPoint(t *testing.T) {
	line := "1, 1.11, 2.22, 333"
	p := NewPoint(line)
	if p.Id_ride != 1 || int(math.Round(p.Coord.Lat*100)) != 111 || int(math.Round(p.Coord.Lon*100)) != 222 || p.Ts != 333 {
		t.Fatalf("Expected point %s but got %s", line, p)
	}
}

func TestWorker(t *testing.T) {
	drive := []string{
		"1,37.966660,23.728308,1405594957",
		"1,37.966627,23.728263,1405594966",
		"1,37.966625,23.728263,1405594974",
		"1,37.966613,23.728375,1405594984",
		"1,37.966203,23.728597,1405594992",
		"1,37.966195,23.728613,1405595001",
		"1,37.966195,23.728613,1405595009",
		"1,37.966195,23.728613,1405595017",
		"1,37.966195,23.728613,1405595026",
		"1,37.966195,23.728613,1405595034",
	}

	var wg sync.WaitGroup
	res := make(chan string)
	wg.Add(1)
	go worker(&wg, drive, res)
	have := <-res
	want := "1, 1.56"
	if want != have {
		t.Fatalf("Wanted %s but got %s", want, have)
	}
}

func TestMerger(t *testing.T) {
	path := "deleteme"
	defer os.Remove(path)
	var wg sync.WaitGroup
	res := make(chan string)
	done := make(chan string)
	wg.Add(1)
	want := "test data"
	go merger(&wg, res, done, path)
	res <- want
	done <- "done"
	wg.Wait()
	haveb, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Cannot read from outfile %s", path)
	}
	have := strings.TrimSpace(string(haveb))
	if string(have) != want {
		t.Fatalf("Wanted '%s' but got '%s'", want, have)
	}
}

func TestDriveWorkers(t *testing.T) {
	path := "resources/paths.csv"
	var wg sync.WaitGroup
	res := make(chan string)
	driveWorkers(path, &wg, res)
	have := 0
	want := 9
	select {
	case <-res:
		have += 1
		if have == want {
			break
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("Time out. Wanted %d, only got %d", want, have)
	}
}
