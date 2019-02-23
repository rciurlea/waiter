package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

type service struct {
	address string
	port    uint64
}

func parseArgs() (uint64, []*service, error) {
	if len(os.Args) < 4 || len(os.Args)%2 == 1 {
		return 0, nil, fmt.Errorf("wrong number of args")
	}
	delay, err := strconv.ParseUint(os.Args[1], 10, 64)
	if err != nil {
		return 0, nil, fmt.Errorf("argument parse: %v", err)
	}
	services := []*service{}
	for i := 2; i < len(os.Args); i += 2 {
		s := &service{address: os.Args[i]}
		port, err := strconv.ParseUint(os.Args[i+1], 10, 64)
		if err != nil {
			return 0, nil, fmt.Errorf("argument parse: port: %v", err)
		}
		s.port = port
		services = append(services, s)
	}
	return delay, services, nil
}

func main() {
	delay, services, err := parseArgs()
	if err != nil {
		log.Println("usage: waiter delay-seconds service1 port1 [service2 port2 ...]")
		log.Fatal(err)
	}

	responsive := map[string]bool{}
	results := make(chan string)
	t := time.NewTimer(time.Duration(delay) * time.Second)

	for _, s := range services {
		responsive[s.address] = false
	}
	for _, s := range services {
		go check(s, results)
	}

	for {
		select {
		case <-t.C:
			os.Exit(1)
		case addr := <-results:
			responsive[addr] = true
			if all(responsive) {
				os.Exit(0)
			}
		}
	}
}

func all(services map[string]bool) bool {
	for _, found := range services {
		if !found {
			return false
		}
	}
	return true
}

func check(s *service, results chan<- string) {
	addr := fmt.Sprintf("%s:%d", s.address, s.port)
	for {
		log.Printf("dialing %s", addr)
		_, err := net.Dial("tcp", addr)
		if err == nil {
			log.Printf("%s reachable!", addr)
			results <- s.address
			return
		}
		log.Printf("%s unreachable", addr)
		time.Sleep(time.Second)
	}
}
