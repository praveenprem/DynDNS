package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/razorcorp/dyndns/cloudflare"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ipv4 struct {
	IP string `json:"ip"`
}

func logger(msg interface{}) {
	log.Printf("%#v", msg)
}

func runner() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}()

	logger("Initiating IP update")

	logger("Cloudflare DNS host found")
	var cf = cloudflare.Cloudflare{}
	cf.Proxied = true

	if token, ok := os.LookupEnv("CF_token"); !ok {
		panic(errors.New("cloudflare Token not set"))
	} else {
		logger("Cloudflare token found")
		cf.Token = token
	}

	if domain, ok := os.LookupEnv("Domain"); !ok {
		panic(errors.New("domain name not set"))
	} else {
		logger("Domain: " + domain)
		cf.Domain = domain
	}

	if hostname, ok := os.LookupEnv("Subdomain"); !ok {
		logger(hostname)
		panic(errors.New("subdomain not set"))
	} else {
		cf.Hostname = hostname
	}

	if _, ok := os.LookupEnv("CF_proxy_disabled"); !ok {
		logger("CF_proxy_disabled not set")
	} else {
		cf.Proxied = false
	}

	logger("Fetching new IP address")
	ip := new(ipv4)

	client := new(http.Client)
	if req, err := http.NewRequest(http.MethodGet, "https://api.ipify.org?format=json", nil); err != nil {
		panic(err)
	} else {
		if resp, err := client.Do(req); err != nil {
			panic(err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				panic(resp.Status)
			} else if err := json.NewDecoder(resp.Body).Decode(&ip); err != nil {
				panic(err)
			}
		}
	}

	logger(ip.IP)

	cf.IpAddress = ip.IP

	logger("Updating DNS record")
	if err := cf.UpdateRecordSet(); err != nil {
		panic(err)
	}

	logger("Update completed")
}

func main() {
	interval := flag.Int64("freq", 2, "IP update frequency in hours. Default 2h")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()

	for {
		runner()
		time.Sleep(time.Hour * time.Duration(*interval))
	}
}
