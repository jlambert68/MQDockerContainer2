package mqcore

import (
	"encoding/hex"
	"testing"
)

func TestNewBrowseID(t *testing.T) {
	// newBrowseID should return a 16-byte random value encoded as 32 hex chars.
	id1, err := newBrowseID()
	if err != nil {
		t.Fatalf("newBrowseID error: %v", err)
	}
	id2, err := newBrowseID()
	if err != nil {
		t.Fatalf("newBrowseID error: %v", err)
	}
	if len(id1) != 32 || len(id2) != 32 {
		t.Fatalf("expected 32 hex chars, got %d and %d", len(id1), len(id2))
	}
	if _, err := hex.DecodeString(id1); err != nil {
		t.Fatalf("id1 not hex: %v", err)
	}
	if _, err := hex.DecodeString(id2); err != nil {
		t.Fatalf("id2 not hex: %v", err)
	}
	if id1 == id2 {
		t.Fatalf("expected different browse IDs, got same value")
	}
}

func TestIntAttr(t *testing.T) {
	// intAttr should coerce common integer types and return 0 for missing.
	attrs := map[int32]interface{}{
		1: int32(5),
		2: int(7),
		3: int64(9),
	}
	if got := intAttr(attrs, 1); got != 5 {
		t.Fatalf("intAttr int32 got %d", got)
	}
	if got := intAttr(attrs, 2); got != 7 {
		t.Fatalf("intAttr int got %d", got)
	}
	if got := intAttr(attrs, 3); got != 9 {
		t.Fatalf("intAttr int64 got %d", got)
	}
	if got := intAttr(attrs, 99); got != 0 {
		t.Fatalf("intAttr missing got %d", got)
	}
}

func TestStringAttr(t *testing.T) {
	// stringAttr should handle string/[]byte and return empty for missing.
	attrs := map[int32]interface{}{
		1: "hello",
		2: []byte("world"),
	}
	if got := stringAttr(attrs, 1); got != "hello" {
		t.Fatalf("stringAttr string got %q", got)
	}
	if got := stringAttr(attrs, 2); got != "world" {
		t.Fatalf("stringAttr bytes got %q", got)
	}
	if got := stringAttr(attrs, 99); got != "" {
		t.Fatalf("stringAttr missing got %q", got)
	}
}

func TestGetenv(t *testing.T) {
	// getenv should return env value when set, otherwise default.
	t.Setenv("MQCORE_TEST_ENV", "set")
	if got := getenv("MQCORE_TEST_ENV", "default"); got != "set" {
		t.Fatalf("getenv set got %q", got)
	}
	if got := getenv("MQCORE_TEST_ENV_MISS", "default"); got != "default" {
		t.Fatalf("getenv default got %q", got)
	}
}

func TestGetbool(t *testing.T) {
	// getbool should parse common truthy/falsey values.
	t.Setenv("MQCORE_TEST_BOOL", "true")
	if got := getbool("MQCORE_TEST_BOOL", false); got != true {
		t.Fatalf("getbool true got %v", got)
	}
	t.Setenv("MQCORE_TEST_BOOL", "0")
	if got := getbool("MQCORE_TEST_BOOL", true); got != false {
		t.Fatalf("getbool false got %v", got)
	}
	t.Setenv("MQCORE_TEST_BOOL", "notabool")
	if got := getbool("MQCORE_TEST_BOOL", true); got != true {
		t.Fatalf("getbool default got %v", got)
	}
}
