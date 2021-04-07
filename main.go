package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/praveenprem/dyndns/cloudflare"
	"log"
	"net/http"
	"os"
)

type ipv4 struct {
	IP string `json:"ip"`
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}()

	var cf = cloudflare.Cloudflare{}
	cf.Proxied = true

	if token, ok := os.LookupEnv("CF_token"); !ok {
		panic(errors.New("cloudflare Token not set"))
	} else {
		cf.Token = token
	}

	if domain, ok := os.LookupEnv("Domain"); !ok {
		panic(errors.New("domain name not set"))
	} else {
		cf.Domain = domain
	}

	if hostname, ok := os.LookupEnv("Subdomain"); !ok {
		log.Print(hostname)
		panic(errors.New("subdomain not set"))
	} else {
		cf.Hostname = hostname
	}

	if _, ok := os.LookupEnv("CF_proxy_disabled"); !ok {
		log.Printf("CF_proxy_disabled not set")
	} else {
		cf.Proxied = false
	}

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

	cf.IpAddress = ip.IP

	if err := cf.UpdateRecordSet(); err != nil {
		panic(err)
	}
}
