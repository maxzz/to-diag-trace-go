//go:build windows

package diag

import (
	"testing"
	"time"
)

func TestParseDeletePaths(t *testing.T) {
	path32, path64 := parseDeletePaths(`/DeleteFiles "C:\ProgramData\DP\Tracing" "C:\ProgramData\DP\Tracing32"`)
	if path64 != `C:\ProgramData\DP\Tracing` {
		t.Fatalf("path64 = %q", path64)
	}
	if path32 != `C:\ProgramData\DP\Tracing32` {
		t.Fatalf("path32 = %q", path32)
	}
}

func TestExtractQuotedStrings(t *testing.T) {
	got := extractQuotedStrings(` "a" b "c" `)
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Fatalf("got %#v", got)
	}
}

func TestStringsJoinArgs(t *testing.T) {
	got := stringsJoinArgs([]string{"/DeleteFiles", `C:\a b`, `C:\c`})
	want := `/DeleteFiles "C:\a b" C:\c`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestIso8601Time(t *testing.T) {
	ts := time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC)
	if iso8601Time(ts) != "2024-01-15T12:30:00Z" {
		t.Fatalf("unexpected iso time: %s", iso8601Time(ts))
	}
}
