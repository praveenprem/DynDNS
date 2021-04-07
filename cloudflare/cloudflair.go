package cloudflare

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type (
	Cloudflare struct {
		//Email    string
		Token     string
		Domain    string
		Hostname  string
		IpAddress string
		Proxied   bool
	}

	Cloudflarer interface {
		UpdateRecordSet() error
		listZones() (*Result, error)
		getRecords(zoneID string) (*Result, error)
		request(method, uri string, body io.Reader) (*http.Request, error)
		createRecord(zone Result) (*Result, error)
		updateRecord(record Result) (*Result, error)
	}

	response struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Messages []struct {
			Code        int    `json:"code"`
			Message     string `json:"message"`
			MessageType string `json:"type"`
		} `json:"messages"`
		Result []Result `json:"result"`
	}

	Result struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Status     string `json:"status"`
		ZoneID     string `json:"zone_id"`
		ZoneName   string `json:"zone_name"`
		RecordType string `json:"type"`
		Content    string `json:"content"`
		Proxiable  bool   `json:"proxiable"`
		Proxied    bool   `json:"proxied"`
		Locked     bool   `json:"locked"`
		TTL        int64  `json:"ttl"`
	}

	Record struct {
		RecordType string `json:"type"`
		Name       string `json:"name"`
		Content    string `json:"content"`
		TTL        int64  `json:"ttl"`
		Proxied    bool   `json:"proxied"`
	}

	UpdateResponse struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Messages []struct {
			Code        int    `json:"code"`
			Message     string `json:"message"`
			MessageType string `json:"type"`
		} `json:"messages"`
		Result *Result `json:"result"`
	}
)

var (
	URL    = "https://api.cloudflare.com/client/v4"
	client = new(http.Client)
)

func (c *Cloudflare) UpdateRecordSet() error {
	zone, zErr := c.listZones()
	if zErr != nil {
		return zErr
	}
	record, rErr := c.getRecords(zone.ID)
	if rErr != nil {
		if errors.Is(rErr, E002) {
			result, resultErr := c.createRecord(*zone)
			if resultErr != nil {
				return resultErr
			}
			log.Printf("record %s set to %s", result.Name, result.Content)
			return nil
		}
		return rErr
	}

	log.Println(record.json())

	if record.Content == c.IpAddress {
		log.Printf("IP not changed since last updated")
		return nil
	}

	result, resultErr := c.updateRecord(*record)
	if resultErr != nil {
		return resultErr
	}

	log.Printf("record %s set to %s", result.Name, result.Content)
	return nil
}

func (c *Cloudflare) request(method, uri string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", URL, uri), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	return req, err
}

func (c *Cloudflare) createRecord(zone Result) (*Result, error) {
	log.Print("creating new record")
	response := new(UpdateResponse)

	payload := new(Record)
	payload.Name = fmt.Sprintf("%s.%s", c.Hostname, c.Domain)
	payload.TTL = 1
	payload.Proxied = c.Proxied
	payload.RecordType = "A"
	payload.Content = c.IpAddress

	req, reqErr := c.request(http.MethodPost, fmt.Sprintf("zones/%s/dns_records", zone.ID),
		bytes.NewBuffer(payload.json()))
	if reqErr != nil {
		log.Print(E006)
		return nil, reqErr
	}

	var resp, respErr = client.Do(req)
	if respErr != nil {
		return nil, respErr
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("failed to decode create record. error message: %s", err.Error())
		return nil, E008
	}

	if resp.StatusCode != 200 {
		log.Print(response.Errors)
		return nil, E009
	}

	if response.Result == nil {
		return nil, E005
	}

	return response.Result, nil
}

func (c *Cloudflare) listZones() (*Result, error) {
	log.Print("getting available zones")
	response := new(response)
	req, err := c.request(http.MethodGet, "zones", nil)
	if err != nil {
		log.Print(E006)
		return nil, err
	}

	params := req.URL.Query()
	params.Add("name", c.Domain)
	req.URL.RawQuery = params.Encode()

	var resp, respErr = client.Do(req)
	if respErr != nil {
		return nil, respErr
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("failed to decode zones. error message: %s", err.Error())
		return nil, E001
	}

	if resp.StatusCode != 200 {
		log.Print(response.Errors)
		return nil, E010
	}

	if len(response.Result) == 0 {
		return nil, E002
	}

	matchZone := new(Result)
	for i, result := range response.Result {
		if result.Name == c.Domain {
			matchZone = &response.Result[i]
			break
		}
	}

	return matchZone, nil
}

func (c *Cloudflare) getRecords(zoneID string) (*Result, error) {
	log.Printf("getting DNS records for zone %s", zoneID)
	response := new(response)
	req, reqErr := c.request(http.MethodGet, fmt.Sprintf("zones/%s/dns_records", zoneID), nil)
	if reqErr != nil {
		log.Print(E006)
		return nil, reqErr
	}

	params := req.URL.Query()
	params.Add("name", fmt.Sprintf("%s.%s", c.Hostname, c.Domain))
	req.URL.RawQuery = params.Encode()

	var resp, respErr = client.Do(req)
	if respErr != nil {
		return nil, respErr
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		log.Printf("failed to decode records. error message: %s", err.Error())
		return nil, E003
	}

	if resp.StatusCode != 200 {
		log.Println(response.Errors[0])
		return nil, E007
	}

	if len(response.Result) == 0 {
		return nil, E002
	}

	if len(response.Result) > 1 {
		return nil, E011
	}

	return &response.Result[0], nil
}

func (c *Cloudflare) updateRecord(record Result) (*Result, error) {
	log.Print("updating existing record")
	response := new(UpdateResponse)
	payload := new(Record)
	payload.Name = record.Name
	payload.TTL = record.TTL
	payload.Proxied = c.Proxied
	payload.RecordType = record.RecordType
	payload.Content = c.IpAddress

	req, reqErr := c.request(http.MethodPut, fmt.Sprintf("zones/%s/dns_records/%s", record.ZoneID, record.ID),
		bytes.NewBuffer(payload.json()))
	if reqErr != nil {
		log.Print(E006)
		return nil, reqErr
	}

	var resp, respErr = client.Do(req)
	if respErr != nil {
		return nil, respErr
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("failed to decode record update. error message: %s", err.Error())
		return nil, E004
	}

	if resp.StatusCode != 200 {
		log.Println(response.Errors)
		return nil, E012
	}

	if response.Result == nil {
		return nil, E005
	}

	return response.Result, nil
}

func (r *Record) json() []byte {
	body, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return body
}

func (res *Result) json() string {
	body, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	return string(body)
}
