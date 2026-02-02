package coolify

import "testing"

func TestParseDeploymentLogs_FiltersAndSorts(t *testing.T) {
	raw := `[
  {"output":"second","type":"stdout","timestamp":"t2","hidden":false,"order":2},
  {"output":"","type":"stdout","timestamp":"t1","hidden":false,"order":1},
  {"output":"hidden","type":"stdout","timestamp":"t3","hidden":true,"order":3},
  {"output":"first","type":"stdout","timestamp":"t1","hidden":false,"order":1}
]`

	entries, err := parseDeploymentLogs(raw)
	if err != nil {
		t.Fatalf("parseDeploymentLogs returned error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Message != "first" {
		t.Fatalf("expected first message to be %q, got %q", "first", entries[0].Message)
	}
	if entries[1].Message != "second" {
		t.Fatalf("expected second message to be %q, got %q", "second", entries[1].Message)
	}
}

func TestParseDeploymentLogs_Empty(t *testing.T) {
	entries, err := parseDeploymentLogs("")
	if err != nil {
		t.Fatalf("parseDeploymentLogs returned error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}
