package powerdns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{
		apiURL: strings.TrimRight(cfg.APIURL, "/"),
		apiKey: cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) CreateZone(zone string, nameservers []string, initialRRSets []RRSet) error {
	canonicalZone := ensureTrailingDot(zone)

	canonicalNS := make([]string, len(nameservers))
	for i, ns := range nameservers {
		canonicalNS[i] = ensureTrailingDot(ns)
	}

	body := CreateZoneRequest{
		Name:        canonicalZone,
		Kind:        "Native",
		Nameservers: canonicalNS,
		RRSets:      initialRRSets,
	}

	_, err := c.do("POST", "/api/v1/servers/localhost/zones", body)
	if err != nil {
		return fmt.Errorf("create zone %s: %w", zone, err)
	}
	return nil
}

func (c *Client) DeleteZone(zone string) error {
	canonicalZone := ensureTrailingDot(zone)

	_, err := c.do("DELETE", fmt.Sprintf("/api/v1/servers/localhost/zones/%s", canonicalZone), nil)
	if err != nil {
		return fmt.Errorf("delete zone %s: %w", zone, err)
	}
	return nil
}

func (c *Client) UpsertRecord(zone, name, rrtype, content string, ttl int) error {
	canonicalZone := ensureTrailingDot(zone)
	canonicalName := ensureTrailingDot(name)

	patch := PatchRRSetsRequest{
		RRSets: []RRSet{
			{
				Name:       canonicalName,
				Type:       rrtype,
				TTL:        ttl,
				ChangeType: "REPLACE",
				Records: []Record{
					{Content: content, Disabled: false},
				},
			},
		},
	}

	_, err := c.do("PATCH", fmt.Sprintf("/api/v1/servers/localhost/zones/%s", canonicalZone), patch)
	if err != nil {
		return fmt.Errorf("upsert record %s %s in %s: %w", name, rrtype, zone, err)
	}
	return nil
}

func (c *Client) DeleteRecord(zone, name, rrtype string) error {
	canonicalZone := ensureTrailingDot(zone)
	canonicalName := ensureTrailingDot(name)

	patch := PatchRRSetsRequest{
		RRSets: []RRSet{
			{
				Name:       canonicalName,
				Type:       rrtype,
				ChangeType: "DELETE",
			},
		},
	}

	_, err := c.do("PATCH", fmt.Sprintf("/api/v1/servers/localhost/zones/%s", canonicalZone), patch)
	if err != nil {
		return fmt.Errorf("delete record %s %s in %s: %w", name, rrtype, zone, err)
	}
	return nil
}

func (c *Client) do(method, path string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.apiURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("PowerDNS API %s %s returned %d: %s", method, path, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func ensureTrailingDot(name string) string {
	if !strings.HasSuffix(name, ".") {
		return name + "."
	}
	return name
}
