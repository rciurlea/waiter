package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	timeout       uint
	checkInterval uint
)

type service struct {
	scheme string
	url    string
}

var services []service

func parseArgs() error {
	flag.UintVar(&timeout, "timeout", 60, "service timeout in seconds")
	flag.UintVar(&checkInterval, "interval", 1, "service check period in seconds")
	flag.Parse()
	if checkInterval > timeout {
		return errors.New("timeout needs to be greater than check interval")
	}
	for _, a := range flag.Args() {
		u, err := url.Parse(a)
		if err != nil {
			return err
		}
		if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "tcp" {
			return fmt.Errorf("invalid service scheme")
		}
		services = append(services,
			service{
				scheme: u.Scheme,
				url:    u.String(),
			},
		)
	}
	if len(services) == 0 {
		return fmt.Errorf("must provide at least on service argument")
	}
	return nil
}

func main() {
	err := parseArgs()
	if err != nil {
		fmt.Println(err)
		fmt.Println(usage)
		os.Exit(1)
	}

	responsive := map[string]bool{}
	results := make(chan string)
	t := time.NewTimer(time.Duration(timeout) * time.Second)

	for _, s := range services {
		responsive[s.url] = false
	}
	for _, s := range services {
		switch s.scheme {
		case "http", "https":
			go checkHTTP(s, results)
		case "tcp":
			go checkTCP(s, results)
		}
	}

	for {
		select {
		case <-t.C:
			os.Exit(1)
		case url := <-results:
			responsive[url] = true
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

func checkTCP(s service, results chan<- string) {
	for {
		log.Printf("dial %s", s.url)
		_, err := net.Dial("tcp", strings.TrimPrefix(s.url, "tcp://"))
		if err == nil {
			log.Printf("%s reachable", s.url)
			results <- s.url
			return
		}
		log.Printf("%s unreachable", s.url)
		time.Sleep(time.Duration(checkInterval) * time.Second)
	}
}

func checkHTTP(s service, results chan<- string) {
	for {
		log.Printf("GET %s", s.url)
		resp, err := http.Get(s.url)
		if err != nil {
			goto end
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			log.Printf("%s reachable", s.url)
			results <- s.url
			return
		}
	end:
		log.Printf("%s unreachable", s.url)
		time.Sleep(time.Duration(checkInterval) * time.Second)
	}
}

const usage = `
usage: waiter -timeout seconds -interval seconds http://service.com https://service2:560 tcp://service3:1234
`
