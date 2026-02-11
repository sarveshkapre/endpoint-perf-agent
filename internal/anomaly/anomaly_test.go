package anomaly

import (
	"strings"
	"testing"
)

func TestDetectorFlagsAnomaly(t *testing.T) {
	detector := NewDetector(5, 2.5)
	values := []float64{10, 11, 9, 10, 12, 50}
	var flagged *Anomaly
	for _, v := range values {
		flagged = detector.Check("cpu_percent", v)
	}
	if flagged == nil {
		t.Fatal("expected anomaly to be flagged")
	}
	if flagged.Name != "cpu_percent" {
		t.Fatalf("unexpected metric name: %s", flagged.Name)
	}
}

func TestDetectorIgnoresNormal(t *testing.T) {
	detector := NewDetector(5, 3.0)
	values := []float64{10, 11, 9, 10, 12, 11, 10}
	for _, v := range values {
		if detector.Check("mem_used_percent", v) != nil {
			t.Fatal("did not expect anomaly")
		}
	}
}

func TestDetectorExplainsDirectionForDrops(t *testing.T) {
	detector := NewDetector(5, 2.5)
	values := []float64{10, 11, 9, 10, 12, 0}
	var flagged *Anomaly
	for _, v := range values {
		flagged = detector.Check("cpu_percent", v)
	}
	if flagged == nil {
		t.Fatal("expected anomaly to be flagged")
	}
	if !strings.Contains(flagged.Explanation, "dropped") {
		t.Fatalf("expected explanation to mention a drop, got: %q", flagged.Explanation)
	}
}

func TestCheckStaticThresholdFlagsExceededValue(t *testing.T) {
	a := CheckStaticThreshold("cpu_percent", 90, map[string]float64{"cpu_percent": 80})
	if a == nil {
		t.Fatalf("expected static-threshold anomaly")
	}
	if a.RuleType != RuleTypeStaticThreshold {
		t.Fatalf("expected static_threshold rule type, got %q", a.RuleType)
	}
	if a.Threshold != 80 {
		t.Fatalf("expected threshold 80, got %v", a.Threshold)
	}
}

func TestSelectHigherSeverityPrefersHigherSeverity(t *testing.T) {
	zscore := &Anomaly{RuleType: RuleTypeZScore, Severity: "medium", ZScore: 3.5}
	static := &Anomaly{RuleType: RuleTypeStaticThreshold, Severity: "high", ZScore: 0.5}
	chosen := SelectHigherSeverity(zscore, static)
	if chosen != static {
		t.Fatalf("expected higher severity anomaly to be selected")
	}
}
