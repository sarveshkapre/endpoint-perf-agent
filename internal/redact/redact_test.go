package redact

import "testing"

func TestParseOptional(t *testing.T) {
	if _, err := ParseOptional("nope"); err == nil {
		t.Fatalf("expected error")
	}
	if m, err := ParseOptional(""); err != nil || m != None {
		t.Fatalf("expected none, got %v err=%v", m, err)
	}
	if m, err := ParseOptional("omit"); err != nil || m != Omit {
		t.Fatalf("expected omit, got %v err=%v", m, err)
	}
	if m, err := ParseOptional("hash"); err != nil || m != Hash {
		t.Fatalf("expected hash, got %v err=%v", m, err)
	}
}

func TestRedactOmit(t *testing.T) {
	if got := HostID("host", Omit); got != "" {
		t.Fatalf("expected empty host, got %q", got)
	}
	if got := Labels(map[string]string{"env": "prod"}, Omit); got != nil {
		t.Fatalf("expected nil labels, got %+v", got)
	}
}

func TestRedactHash(t *testing.T) {
	host := HostID("host", Hash)
	if host == "" || host == "host" {
		t.Fatalf("expected hashed host, got %q", host)
	}
	labels := Labels(map[string]string{"env": "prod"}, Hash)
	if labels["env"] == "" || labels["env"] == "prod" {
		t.Fatalf("expected hashed label value, got %+v", labels)
	}
}
