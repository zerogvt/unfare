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

func worker(wg *sync.WaitGroup, drive []string) {
	defer wg.Done()
	var prev_p *Point = nil
	daily_rate := 0.74
	nightly_rate := 1.30
	idle_rate_per_sec := 11.9 / 3600
	fare := 0.0
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
	fare += 1.3 //flag amount
	fmt.Println(prev_p.Id_ride, fare)
}

func main() {
	var wg sync.WaitGroup
	file, err := os.Open("paths.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	curr_num := 1
	curr_drive := []string{}
	//read line by line
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
			wg.Add(1)
			go worker(&wg, curr_drive)
			curr_num = num
			curr_drive = nil
			curr_drive = append(curr_drive, line)
		}
	}
	//don't forget the last drive
	wg.Add(1)
	go worker(&wg, curr_drive)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	wg.Wait()
}
