package stats

import (
	"testing"

	"shortlink/internal/channel"
)

// Documents Phase A handoff expectations for assistant tools.
func TestPhaseAContractConstants(t *testing.T) {
	if ExpectedYesterdayTopChannel() != channel.SMS {
		t.Fatal(ExpectedYesterdayTopChannel())
	}
	codes := DemoShortCodes()
	if len(codes) != 3 {
		t.Fatalf("%v", codes)
	}
	if VisitorHash("1.2.3.4", "ua") == VisitorHash("1.2.3.4", "ub") {
		t.Fatal("ua must affect hash")
	}
	if len(VisitorHash("a", "b")) != 64 {
		t.Fatal("sha256 hex length")
	}
}
