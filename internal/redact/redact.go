package redact

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type Mode int

const (
	None Mode = iota
	Omit
	Hash
)

func ParseOptional(s string) (Mode, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "":
		return None, nil
	case "omit":
		return Omit, nil
	case "hash":
		return Hash, nil
	default:
		return None, fmt.Errorf("unknown redact mode: %s (expected omit|hash)", s)
	}
}

func HostID(hostID string, mode Mode) string {
	switch mode {
	case Omit:
		return ""
	case Hash:
		if strings.TrimSpace(hostID) == "" {
			return ""
		}
		return "hash:" + shortHash(hostID)
	default:
		return hostID
	}
}

func Labels(labels map[string]string, mode Mode) map[string]string {
	if len(labels) == 0 {
		return nil
	}
	switch mode {
	case Omit:
		return nil
	case Hash:
		out := make(map[string]string, len(labels))
		for k, v := range labels {
			if strings.TrimSpace(k) == "" {
				continue
			}
			if strings.TrimSpace(v) == "" {
				out[k] = ""
				continue
			}
			out[k] = "hash:" + shortHash(v)
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		// Clone to avoid accidental mutation of shared maps.
		out := make(map[string]string, len(labels))
		for k, v := range labels {
			if strings.TrimSpace(k) == "" {
				continue
			}
			out[k] = v
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}
}

func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	hexed := hex.EncodeToString(sum[:])
	if len(hexed) > 12 {
		return hexed[:12]
	}
	return hexed
}
