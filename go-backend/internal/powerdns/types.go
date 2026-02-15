package powerdns

type Zone struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Kind           string   `json:"kind"`
	DNSSec         bool     `json:"dnssec"`
	Nameservers    []string `json:"nameservers"`
	Serial         int64    `json:"serial"`
	NotifiedSerial int64    `json:"notified_serial"`
}

type RRSet struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	TTL        int      `json:"ttl"`
	ChangeType string   `json:"changetype"`
	Records    []Record `json:"records"`
}

type Record struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`
}

type CreateZoneRequest struct {
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	Nameservers []string `json:"nameservers"`
	RRSets      []RRSet  `json:"rrsets,omitempty"`
}

type PatchRRSetsRequest struct {
	RRSets []RRSet `json:"rrsets"`
}
