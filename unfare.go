package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// represents a point in a drive
type Point struct {
	Id_ride int
	Coord   Coordinates
	Ts      int64
}

func (p *Point) String() string {
	return fmt.Sprintf("%v, %v, %v, %v", p.Id_ride, p.Coord.Lat, p.Coord.Lon, p.Ts)
}

// calc velocity km/h based on hameshine dist
func (p1 *Point) velocity(p2 *Point) float64 {
	vel := 3600 * Distance(p1.Coord, p2.Coord) / math.Abs(float64(p1.Ts-p2.Ts))
	return vel
}

// calc fare for prev_p to p
func (p1 *Point) fareTo(p2 *Point) (fare float64, err error) {
	daily_rate := 0.74
	nightly_rate := 1.30
	idle_rate_per_sec := 11.9 / 3600
	vel := p1.velocity(p2)
	if vel > 100.0 {
		return 0.0, errors.New("Outlier point")
	}
	// calc fare based on the km travelled and time of day
	dist := Distance(p1.Coord, p2.Coord)
	t_start := time.Unix(p1.Ts, 0)
	t_end := time.Unix(p2.Ts, 0)
	if vel > 10 {
		if t_start.Hour() < 5 {
			fare = nightly_rate * dist
		} else {
			fare = daily_rate * dist
		}
	} else {
		fare = idle_rate_per_sec * float64(t_end.Sub(t_start).Seconds())
	}
	return fare, nil
}

// constructor for a new point
func NewPoint(line string) *Point {
	tokens := strings.Split(line, ",")
	if len(tokens) != 4 {
		return nil
	}
	var err error
	p := new(Point)
	tok := strings.TrimSpace(tokens[0])
	p.Id_ride, err = strconv.Atoi(tok)
	if err != nil {
		return nil
	}
	tok = strings.TrimSpace(tokens[1])
	p.Coord.Lat, err = strconv.ParseFloat(tok, 32)
	if err != nil {
		return nil
	}
	tok = strings.TrimSpace(tokens[2])
	p.Coord.Lon, err = strconv.ParseFloat(tok, 32)
	if err != nil {
		return nil
	}
	tok = strings.TrimSpace(tokens[3])
	p.Ts, err = strconv.ParseInt(tok, 10, 64)
	if err != nil {
		return nil
	}
	return p
}

// worker calculates total fare for a complete drive
func worker(wg *sync.WaitGroup, drive []string, res chan string) {
	// inform waitgroup when we're done
	defer wg.Done()
	var prev_p *Point = nil
	// fare starts with the flag
	fare := 1.3
	// trace the drive point to point
	// weed out outlier points
	// calc fare point to point
	for _, line := range drive {
		p := NewPoint(line)
		if p == nil {
			continue
		}
		// first valid point marks the begining
		if prev_p == nil {
			prev_p = p
			continue
		}
		// get p2p fare if any
		p2pfare, err := prev_p.fareTo(p)
		if err != nil {
			continue
		}
		// add to total fare and update current pos
		fare += p2pfare
		prev_p = p
	}
	// round up to 2 decimal points
	fare = math.Round(fare*100) / 100
	// sent our result to the merger goroutine
	res <- fmt.Sprint(prev_p.Id_ride, ", ", fare)
}

// merger waits to hear results over channel res untill something is sent over channel done
// appends each result line to the output file in path
// informs its waitgroup when done
func merger(wg *sync.WaitGroup, res chan string, done chan string, path string) {
	defer wg.Done()
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Cannot create file ", path)
		return
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	for {
		select {
		case r := <-res:
			fmt.Println(r)
			fmt.Fprintln(w, r)
		case <-done:
			w.Flush()
			return
		}
	}
}

// start reading in input data
// When we complete a ride start a worker goroutine
// to weed out outlier points and calculate fare
func driveWorkers(infile string, wg *sync.WaitGroup, res chan string) {
	file, err := os.Open(infile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	curr_num := 1
	curr_drive := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, ",")
		num, err := strconv.Atoi(tokens[0])
		if err != nil {
			continue
		}
		if num == curr_num {
			curr_drive = append(curr_drive, line)
		} else {
			// a complete drive was read in. Start a worker on it
			wg.Add(1)
			go worker(wg, curr_drive, res)
			curr_num = num
			curr_drive = nil
			curr_drive = append(curr_drive, line)
		}
	}
	//don't forget the last drive
	wg.Add(1)
	go worker(wg, curr_drive, res)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) != 2 {
		fmt.Println("Usage: unfare [input_data_file] [output_file]")
		return
	}
	infile := argsWithoutProg[0]
	outfile := argsWithoutProg[1]

	var worker_wg sync.WaitGroup
	var merger_wg sync.WaitGroup
	res := make(chan string)
	done := make(chan string)

	// start the results merge goroutine
	// it will wait workers results untill notified that
	// all of them are finished
	go merger(&merger_wg, res, done, outfile)
	merger_wg.Add(1)

	//let the driver start reading through the file
	//and assign work to workers
	driveWorkers(infile, &worker_wg, res)

	// wait all the workers to finish
	worker_wg.Wait()

	// signal to merger that all workers are done
	// and wait for it to finish flushing content
	done <- "done"
	merger_wg.Wait()
}
