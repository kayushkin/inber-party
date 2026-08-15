package inber

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestParseTimestampInAnyKnownFormat(t *testing.T) {
	cases := []struct {
		name     string
		value    string
		wantHour int
		wantOK   bool
	}{
		// What a bare DATETIME column comes back as, once the driver has turned
		// it into a time.Time and it has been scanned back out into a string.
		{"RFC 3339 in UTC", "2026-08-14T02:15:00Z", 2, true},
		{"RFC 3339 with fractional seconds", "2026-08-14T02:15:00.123456789Z", 2, true},
		// What the same column comes back as when read through an expression.
		{"space separated", "2026-08-14 02:15:00", 2, true},
		// What the HTTP API sends.
		{"RFC 3339 with a numeric offset", "2026-08-14T02:15:00-07:00", 2, true},

		{"empty", "", 0, false},
		{"not a timestamp", "yesterday", 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseTimestampInAnyKnownFormat(tc.value)
			if ok != tc.wantOK {
				t.Fatalf("parseTimestampInAnyKnownFormat(%q) recognised = %v, want %v", tc.value, ok, tc.wantOK)
			}
			if ok && got.Hour() != tc.wantHour {
				t.Errorf("parseTimestampInAnyKnownFormat(%q) hour = %d, want %d", tc.value, got.Hour(), tc.wantHour)
			}
		})
	}
}

// TestTheDriverReturnsTwoTimestampFormatsFromOneColumn pins the behaviour that
// made the Night Owl achievement unreachable for as long as it was.
//
// The sqlite driver converts a column whose declared type is DATETIME into a
// time.Time, so reading it back as a string gives RFC 3339. Read the very same
// column through an aggregate and the declared type is gone, so the stored text
// arrives untouched. Both formats therefore come out of one table, and code that
// parsed against only the second one worked or failed depending on nothing more
// than the shape of the query above it.
func TestTheDriverReturnsTwoTimestampFormatsFromOneColumn(t *testing.T) {
	db, err := sql.Open("sqlite3", createAchievementGatewayDB(t))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var bare, aggregated sql.NullString
	if err := db.QueryRow(`SELECT started_at FROM requests WHERE id = 2`).Scan(&bare); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT MAX(started_at) FROM requests WHERE id = 2`).Scan(&aggregated); err != nil {
		t.Fatal(err)
	}

	if bare.String == aggregated.String {
		t.Fatalf("expected the driver to render one column two ways, got %q from both", bare.String)
	}
	for label, value := range map[string]string{"bare column": bare.String, "aggregate": aggregated.String} {
		parsed, ok := parseTimestampInAnyKnownFormat(value)
		if !ok {
			t.Errorf("%s returned %q, which no known timestamp format parses", label, value)
			continue
		}
		if parsed.Hour() != 2 {
			t.Errorf("%s returned %q, parsed to hour %d, want 2", label, value, parsed.Hour())
		}
	}
}

// TestGetAgentsParsesLastActive pins the other site that reads a timestamp out
// of the gateway database. It is correct today only because its query wraps the
// column in MAX(); nothing said so, and nothing noticed when the parse failed —
// a nil time silently becomes full energy rather than an error.
func TestGetAgentsParsesLastActive(t *testing.T) {
	store, err := NewStore("", createTestGatewayDB(t), "")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	agents, err := store.GetAgents()
	if err != nil {
		t.Fatal(err)
	}
	if len(agents) == 0 {
		t.Fatal("no agents")
	}
	for i := range agents {
		if agents[i].LastActive == nil {
			t.Errorf("agent %s has a nil LastActive, so its timestamp did not parse", agents[i].ID)
		}
	}
}
