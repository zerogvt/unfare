package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Point struct {
	Id_ride int
	Coord   Coordinates
	Ts      int64
}

func (p1 *Point) velocity(p2 *Point) float64 {
	return 3600 * Distance(p1.Coord, p2.Coord) / math.Abs(float64(p1.Ts-p2.Ts))
}

func NewPoint(line string) *Point {
	tokens := strings.Split(line, ",")
	if len(tokens) != 4 {
		return nil
	}
	var err error
	p := new(Point)
	p.Id_ride, err = strconv.Atoi(tokens[0])
	if err != nil {
		return nil
	}
	p.Coord.Lat, err = strconv.ParseFloat(tokens[1], 32)
	if err != nil {
		return nil
	}
	p.Coord.Lon, err = strconv.ParseFloat(tokens[2], 32)
	if err != nil {
		return nil
	}
	p.Ts, err = strconv.ParseInt(tokens[3], 10, 64)
	if err != nil {
		return nil
	}
	return p
}

func worker(wg *sync.WaitGroup, drive []string, res chan string) {
	// inform waitgroup when we're done
	defer wg.Done()
	var prev_p *Point = nil
	daily_rate := 0.74
	nightly_rate := 1.30
	idle_rate_per_sec := 11.9 / 3600
	fare := 0.0

	// trace the drive point to point
	// weed out outlier points
	// calc fare point to point
	for _, line := range drive {
		p := NewPoint(line)
		if p == nil {
			continue
		}
		if prev_p == nil {
			prev_p = p
			continue
		}
		vel := p.velocity(prev_p)
		if vel > 100.0 {
			continue
		}
		// calc fare for prev_p to p
		dist := Distance(prev_p.Coord, p.Coord)
		t_start := time.Unix(prev_p.Ts, 0)
		t_end := time.Unix(p.Ts, 0)
		if vel > 10 {
			if t_start.Hour() < 5 {
				fare += nightly_rate * dist
			} else {
				fare += daily_rate * dist
			}
		} else {
			fare += idle_rate_per_sec * float64(t_end.Sub(t_start).Seconds())
		}
		prev_p = p
	}
	// add in flag amount
	fare += 1.3
	fare = math.Round(fare*100) / 100
	// sent our result to the merger goroutine
	res <- fmt.Sprint(prev_p.Id_ride, ", ", fare)
}

// waits to hear results over channel res untill something is sent over channel done
// appends each result line to the output file in path
// informs its waitgroup when done
func merge(wg *sync.WaitGroup, res chan string, done chan string, path string) {
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
	file, err := os.Open(infile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	curr_num := 1
	curr_drive := []string{}

	// start the results merge goroutine
	go merge(&merger_wg, res, done, outfile)
	merger_wg.Add(1)

	// start reading in input data
	// When we complete a ride start a worker goroutine
	// to weed out outlier points and calculate fare
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
			worker_wg.Add(1)
			go worker(&worker_wg, curr_drive, res)
			curr_num = num
			curr_drive = nil
			curr_drive = append(curr_drive, line)
		}
	}
	//don't forget the last drive
	worker_wg.Add(1)
	go worker(&worker_wg, curr_drive, res)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	// wait all the workers to finish
	worker_wg.Wait()
	// signal to merger that all workers are done
	done <- "done"
	merger_wg.Wait()
}
