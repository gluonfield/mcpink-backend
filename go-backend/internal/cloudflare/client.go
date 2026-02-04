package cloudflare

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"
)

type Client struct {
	api    *cf.API
	config Config
	logger *slog.Logger
}

func NewClient(config Config, logger *slog.Logger) (*Client, error) {
	if config.APIToken == "" {
		return nil, fmt.Errorf("cloudflare: APIToken is required")
	}
	if config.ZoneID == "" {
		return nil, fmt.Errorf("cloudflare: ZoneID is required")
	}
	if config.BaseDomain == "" {
		return nil, fmt.Errorf("cloudflare: BaseDomain is required")
	}

	api, err := cf.NewWithAPIToken(config.APIToken)
	if err != nil {
		return nil, fmt.Errorf("cloudflare: failed to create client: %w", err)
	}

	return &Client{
		api:    api,
		config: config,
		logger: logger,
	}, nil
}

func (c *Client) Config() Config {
	return c.config
}

type DNSRecord struct {
	ID         string
	Name       string
	Type       string
	Content    string
	Proxied    bool
	TTL        int
	FullDomain string
}

func (c *Client) CreateARecord(ctx context.Context, subdomain, targetIP string) (*DNSRecord, error) {
	fullDomain := fmt.Sprintf("%s.%s", subdomain, c.config.BaseDomain)

	proxied := true
	record, err := c.api.CreateDNSRecord(ctx, cf.ZoneIdentifier(c.config.ZoneID), cf.CreateDNSRecordParams{
		Type:    "A",
		Name:    fullDomain,
		Content: targetIP,
		Proxied: &proxied,
		TTL:     1, // Auto TTL when proxied
	})
	if err != nil {
		return nil, fmt.Errorf("cloudflare: failed to create A record: %w", err)
	}

	c.logger.Info("created DNS record",
		"subdomain", subdomain,
		"full_domain", fullDomain,
		"target_ip", targetIP,
		"record_id", record.ID)

	return &DNSRecord{
		ID:         record.ID,
		Name:       subdomain,
		Type:       record.Type,
		Content:    record.Content,
		Proxied:    *record.Proxied,
		TTL:        record.TTL,
		FullDomain: fullDomain,
	}, nil
}

func (c *Client) DeleteDNSRecord(ctx context.Context, recordID string) error {
	err := c.api.DeleteDNSRecord(ctx, cf.ZoneIdentifier(c.config.ZoneID), recordID)
	if err != nil {
		return fmt.Errorf("cloudflare: failed to delete DNS record: %w", err)
	}

	c.logger.Info("deleted DNS record", "record_id", recordID)
	return nil
}

func (c *Client) GetDNSRecordByName(ctx context.Context, subdomain string) (*DNSRecord, error) {
	fullDomain := fmt.Sprintf("%s.%s", subdomain, c.config.BaseDomain)

	records, _, err := c.api.ListDNSRecords(ctx, cf.ZoneIdentifier(c.config.ZoneID), cf.ListDNSRecordsParams{
		Name: fullDomain,
		Type: "A",
	})
	if err != nil {
		return nil, fmt.Errorf("cloudflare: failed to list DNS records: %w", err)
	}

	if len(records) == 0 {
		return nil, nil
	}

	r := records[0]
	proxied := false
	if r.Proxied != nil {
		proxied = *r.Proxied
	}

	return &DNSRecord{
		ID:         r.ID,
		Name:       subdomain,
		Type:       r.Type,
		Content:    r.Content,
		Proxied:    proxied,
		TTL:        r.TTL,
		FullDomain: fullDomain,
	}, nil
}

func (c *Client) UpdateARecord(ctx context.Context, recordID, subdomain, targetIP string) (*DNSRecord, error) {
	fullDomain := fmt.Sprintf("%s.%s", subdomain, c.config.BaseDomain)

	proxied := true
	record, err := c.api.UpdateDNSRecord(ctx, cf.ZoneIdentifier(c.config.ZoneID), cf.UpdateDNSRecordParams{
		ID:      recordID,
		Type:    "A",
		Name:    fullDomain,
		Content: targetIP,
		Proxied: &proxied,
		TTL:     1,
	})
	if err != nil {
		return nil, fmt.Errorf("cloudflare: failed to update A record: %w", err)
	}

	c.logger.Info("updated DNS record",
		"subdomain", subdomain,
		"full_domain", fullDomain,
		"target_ip", targetIP,
		"record_id", record.ID)

	return &DNSRecord{
		ID:         record.ID,
		Name:       subdomain,
		Type:       record.Type,
		Content:    record.Content,
		Proxied:    *record.Proxied,
		TTL:        record.TTL,
		FullDomain: fullDomain,
	}, nil
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9-]`)

func SanitizeSubdomain(name string) string {
	name = strings.ToLower(name)
	name = nonAlphanumeric.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if len(name) > 63 {
		name = name[:63]
	}
	return name
}
