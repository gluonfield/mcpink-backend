package dns

import (
	"context"
	"fmt"

	"github.com/augustdev/autoclip/internal/powerdns"
)

func (a *Activities) CreateZone(ctx context.Context, input CreateZoneInput) error {
	a.logger.Info("CreateZone", "zone", input.Zone)

	wildcardRRSet := powerdns.RRSet{
		Name:       "*." + input.Zone + ".",
		Type:       "A",
		TTL:        60,
		ChangeType: "REPLACE",
		Records: []powerdns.Record{
			{Content: input.IngressIP, Disabled: false},
		},
	}

	if err := a.pdns.CreateZone(input.Zone, input.Nameservers, []powerdns.RRSet{wildcardRRSet}); err != nil {
		return fmt.Errorf("create zone: %w", err)
	}
	return nil
}

func (a *Activities) DeleteZone(ctx context.Context, input DeleteZoneInput) error {
	a.logger.Info("DeleteZone", "zone", input.Zone)

	if err := a.pdns.DeleteZone(input.Zone); err != nil {
		return fmt.Errorf("delete zone: %w", err)
	}
	return nil
}

func (a *Activities) UpsertRecord(ctx context.Context, input UpsertRecordInput) error {
	a.logger.Info("UpsertRecord", "zone", input.Zone, "name", input.Name, "type", input.Type)

	if err := a.pdns.UpsertRecord(input.Zone, input.Name, input.Type, input.Content, input.TTL); err != nil {
		return fmt.Errorf("upsert record: %w", err)
	}
	return nil
}

func (a *Activities) DeleteRecord(ctx context.Context, input DeleteRecordInput) error {
	a.logger.Info("DeleteRecord", "zone", input.Zone, "name", input.Name, "type", input.Type)

	if err := a.pdns.DeleteRecord(input.Zone, input.Name, input.Type); err != nil {
		return fmt.Errorf("delete record: %w", err)
	}
	return nil
}
