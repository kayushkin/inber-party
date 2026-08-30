package dailyquests

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kayushkin/inber-party/internal/db"

	_ "github.com/mattn/go-sqlite3"
)

// What these tests pin.
//
// GetActiveDailyQuests skipped a task row it could not scan and returned the shorter slice
// with a nil error, and it never checked rows.Err(). A row that could not be read and a row
// that was not there produced the same answer.
//
// The reason that matters more here than at a bare list endpoint: GetQuestStats publishes
// active_daily_quests as a SQL COUNT(*) over the same predicate. So the dropped row IS
// reported as a count -- by the neighbouring endpoint -- and the two answers disagreed with
// nothing on either of them saying so.
//
// ⚠️ Why there is no SQLite arm on GetActiveDailyQuests itself, stated rather than left as a
// gap for a reader to discover. The shipped statement is PostgreSQL-only:
//
//	sqlite3: near "'1 day'": syntax error
//
// measured directly, so the fixture route the agent-row tests use is not available at this
// site. What stands in for it is TestTheShippedQuerySelectsWhatTheScanReads below, which ties
// activeDailyQuestsQuery's column list to the Scan destinations -- the one property the
// positional scan depends on and the one a helper-only test would otherwise leave unpinned.

// stubDailyQuestRowCursor drives scanDailyQuestRows through outcomes a real driver gives no
// reliable way to produce on demand, in particular a failure of the iteration itself.
//
// scanRow, when set, is called per row so a test can write real values into the destinations
// and check where each one lands.
type stubDailyQuestRowCursor struct {
	remaining int
	scanRow   func(dest ...any) error
	scanErr   error
	iterErr   error
}

func (c *stubDailyQuestRowCursor) Next() bool {
	if c.remaining == 0 {
		return false
	}
	c.remaining--
	return true
}

func (c *stubDailyQuestRowCursor) Scan(dest ...any) error {
	if c.scanRow != nil {
		return c.scanRow(dest...)
	}
	return c.scanErr
}

func (c *stubDailyQuestRowCursor) Err() error { return c.iterErr }

// TestEveryRowReadableReportsZeroUnreadableQuests is the known-negative arm. Without it a
// scanDailyQuestRows that returned a constant 1 would pass the positive arm below.
func TestEveryRowReadableReportsZeroUnreadableQuests(t *testing.T) {
	cursor := &stubDailyQuestRowCursor{remaining: 2, scanRow: func(dest ...any) error { return nil }}

	quests, unreadableRows, err := scanDailyQuestRows(cursor)
	if err != nil {
		t.Fatalf("scanDailyQuestRows: %v", err)
	}
	if len(quests) != 2 {
		t.Errorf("read %d quests, want 2", len(quests))
	}
	if unreadableRows != 0 {
		t.Errorf("reported %d unreadable rows, want 0 -- every row in this cursor scans", unreadableRows)
	}
}

// TestAnUnreadableQuestRowIsCountedNotErased is the defect this change repairs. The row is
// still skipped; what changed is that the caller is told it was skipped.
func TestAnUnreadableQuestRowIsCountedNotErased(t *testing.T) {
	// Rows 1 and 3 scan; row 2 is the NULL the schema permits. Only id is written, which is
	// enough to tell the surviving rows apart.
	row := 0
	cursor := &stubDailyQuestRowCursor{remaining: 3, scanRow: func(dest ...any) error {
		row++
		if row == 2 {
			return errors.New("converting NULL to string is unsupported")
		}
		id, ok := dest[0].(*int)
		if !ok {
			return errors.New("first scan destination is not an *int")
		}
		*id = row
		return nil
	}}

	quests, unreadableRows, err := scanDailyQuestRows(cursor)
	if err != nil {
		t.Fatalf("scanDailyQuestRows: %v", err)
	}
	if len(quests) != 2 {
		t.Fatalf("read %d quests, want 2 -- the unscannable row must still be skipped", len(quests))
	}
	if unreadableRows != 1 {
		t.Errorf("reported %d unreadable rows, want 1 -- the skipped row is the whole finding, and a 0 here is the old behaviour where it vanished", unreadableRows)
	}
	for _, quest := range quests {
		if quest.ID == 2 {
			t.Errorf("quest id 2 was returned, but its row did not scan")
		}
	}

	// The distinguishing assertion. Two quests came back either way; only the count separates
	// "there are two daily quests" from "there are three and one could not be read".
	if len(quests)+unreadableRows != 3 {
		t.Errorf("quests read (%d) + rows unreadable (%d) = %d, want 3 -- the two numbers together must account for every row the database returned", len(quests), unreadableRows, len(quests)+unreadableRows)
	}
}

// TestAQuestIterationFailureIsReturnedNotCounted pins the second silent drop at this site.
// If row iteration fails part-way the loop simply ends: every remaining row is lost, and how
// many were lost is not knowable because the driver stopped wherever it failed. Counting them
// would be a guess and returning the short slice with a nil error would present it as
// complete, so the error is returned instead.
func TestAQuestIterationFailureIsReturnedNotCounted(t *testing.T) {
	iterationFailure := errors.New("connection reset mid-scan")
	cursor := &stubDailyQuestRowCursor{
		remaining: 2,
		scanRow:   func(dest ...any) error { return nil },
		iterErr:   iterationFailure,
	}

	quests, unreadableRows, err := scanDailyQuestRows(cursor)
	if err == nil {
		t.Fatal("scanDailyQuestRows returned a nil error after rows.Err() failed -- a truncated list presented as complete is the drop this repair exists to stop")
	}
	if !errors.Is(err, iterationFailure) {
		t.Errorf("returned error %v does not wrap the iteration failure", err)
	}
	if quests != nil {
		t.Errorf("returned %d quests alongside the error, want none -- a partial list invites the caller to use it", len(quests))
	}
	if unreadableRows != 0 {
		t.Errorf("returned unreadableRows = %d alongside the error, want 0 -- the number lost is not knowable here, so any count would be invented", unreadableRows)
	}
}

// TestQuestScanFailuresAreCountedNotReturnedAsAnError is the discriminating negative for the
// test above: a scan failure and an iteration failure must not be handled the same way. A
// repair that returned an error on any failure would pass the iteration arm and break the
// endpoint on a single NULL.
func TestQuestScanFailuresAreCountedNotReturnedAsAnError(t *testing.T) {
	cursor := &stubDailyQuestRowCursor{remaining: 3, scanErr: errors.New("converting NULL to int is unsupported")}

	quests, unreadableRows, err := scanDailyQuestRows(cursor)
	if err != nil {
		t.Fatalf("scanDailyQuestRows returned %v; a row that will not scan is counted, not fatal", err)
	}
	if len(quests) != 0 {
		t.Errorf("read %d quests, want 0", len(quests))
	}
	if unreadableRows != 3 {
		t.Errorf("unreadableRows = %d, want 3", unreadableRows)
	}
}

// TestEveryQuestColumnLandsInItsOwnField is the positional-binding arm. Swap two columns of
// the same type in activeDailyQuestsQuery -- xp_reward and progress are both integers -- and
// the query still runs, every row still scans, the column count still agrees, and every other
// test in this file still passes. The values simply land in each other's fields. Giving each
// column a value no other column has and checking where it arrives is what catches it.
func TestEveryQuestColumnLandsInItsOwnField(t *testing.T) {
	createdAt := time.Date(2026, 8, 30, 4, 5, 6, 0, time.UTC)
	startedAt := time.Date(2026, 8, 30, 7, 8, 9, 0, time.UTC)
	completedAt := time.Date(2026, 8, 30, 10, 11, 12, 0, time.UTC)
	agentID, partyID := 71, 92

	// Written positionally, in the order activeDailyQuestsQuery selects. Every value is
	// unique across the row, so a column landing in the wrong field cannot be masked by
	// another column happening to hold the same value.
	values := []any{
		int(13),         // id
		"[DAILY] Forge", // name
		"a description", // description
		"hard",          // difficulty
		int(37),         // xp_reward
		"in_progress",   // status
		&agentID,        // assigned_agent_id
		&partyID,        // assigned_party_id
		int(58),         // progress
		createdAt,       // created_at
		&startedAt,      // started_at
		&completedAt,    // completed_at
	}

	cursor := &stubDailyQuestRowCursor{remaining: 1, scanRow: func(dest ...any) error {
		if len(dest) != len(values) {
			t.Fatalf("Scan received %d destinations, but this case writes %d values -- the SELECT list and this case are out of step", len(dest), len(values))
		}
		for i, value := range values {
			switch target := dest[i].(type) {
			case *int:
				v, ok := value.(int)
				if !ok {
					t.Fatalf("destination %d is *int but this case holds %T", i, value)
				}
				*target = v
			case *string:
				v, ok := value.(string)
				if !ok {
					t.Fatalf("destination %d is *string but this case holds %T", i, value)
				}
				*target = v
			case **int:
				v, ok := value.(*int)
				if !ok {
					t.Fatalf("destination %d is **int but this case holds %T", i, value)
				}
				*target = v
			case *time.Time:
				v, ok := value.(time.Time)
				if !ok {
					t.Fatalf("destination %d is *time.Time but this case holds %T", i, value)
				}
				*target = v
			case **time.Time:
				v, ok := value.(*time.Time)
				if !ok {
					t.Fatalf("destination %d is **time.Time but this case holds %T", i, value)
				}
				*target = v
			default:
				t.Fatalf("destination %d has unhandled type %T", i, dest[i])
			}
		}
		return nil
	}}

	quests, unreadableRows, err := scanDailyQuestRows(cursor)
	if err != nil {
		t.Fatalf("scanDailyQuestRows: %v", err)
	}
	if unreadableRows != 0 {
		t.Fatalf("unreadableRows = %d, want 0", unreadableRows)
	}
	if len(quests) != 1 {
		t.Fatalf("read %d quests, want 1", len(quests))
	}
	quest := quests[0]

	for _, check := range []struct {
		column string
		got    any
		want   any
	}{
		{"id", quest.ID, 13},
		{"name", quest.Name, "[DAILY] Forge"},
		{"description", quest.Description, "a description"},
		{"difficulty", quest.Difficulty, "hard"},
		{"xp_reward", quest.XPReward, 37},
		{"status", quest.Status, "in_progress"},
		{"progress", quest.Progress, 58},
		{"created_at", quest.CreatedAt, createdAt},
	} {
		if check.got != check.want {
			t.Errorf("column %s landed as %v, want %v -- the SELECT list and the Scan destinations are out of step", check.column, check.got, check.want)
		}
	}
	if quest.AssignedAgentID == nil || *quest.AssignedAgentID != 71 {
		t.Errorf("assigned_agent_id landed as %v, want 71", quest.AssignedAgentID)
	}
	if quest.AssignedPartyID == nil || *quest.AssignedPartyID != 92 {
		t.Errorf("assigned_party_id landed as %v, want 92", quest.AssignedPartyID)
	}
	if quest.StartedAt == nil || !quest.StartedAt.Equal(startedAt) {
		t.Errorf("started_at landed as %v, want %v", quest.StartedAt, startedAt)
	}
	if quest.CompletedAt == nil || !quest.CompletedAt.Equal(completedAt) {
		t.Errorf("completed_at landed as %v, want %v", quest.CompletedAt, completedAt)
	}
}

// TestTheShippedQuerySelectsWhatTheScanReads is what stands in for the SQLite arm the agent
// tests have, and it is the only test here that reads the shipped SQL at all.
//
// scanDailyQuestRows binds by position, so activeDailyQuestsQuery's SELECT list and the Scan
// destination list are one statement written twice. Comparing their LENGTHS is not enough:
// swapping two same-typed columns in the SELECT -- xp_reward and progress are both integers --
// keeps the lengths equal, keeps the query runnable, keeps every row scanning, and lands each
// value in the other's field. Measured: that mutation was MISSED by every other test in this
// file, TestEveryQuestColumnLandsInItsOwnField included, because that one drives the stub
// cursor and so never reads the query.
//
// So this walks the SELECT list by position and checks that the destination at that position
// is the field the column belongs in, by writing a marker through the scan and finding it.
func TestTheShippedQuerySelectsWhatTheScanReads(t *testing.T) {
	selected := selectedColumns(t, activeDailyQuestsQuery)

	// fieldHoldingColumn names, per column, the db.Task field that column must land in. It is
	// the pairing the positional scan asserts silently and this test asserts out loud.
	fieldHoldingColumn := map[string]func(db.Task) any{
		"id":                func(q db.Task) any { return q.ID },
		"name":              func(q db.Task) any { return q.Name },
		"description":       func(q db.Task) any { return q.Description },
		"difficulty":        func(q db.Task) any { return q.Difficulty },
		"xp_reward":         func(q db.Task) any { return q.XPReward },
		"status":            func(q db.Task) any { return q.Status },
		"assigned_agent_id": func(q db.Task) any { return derefInt(q.AssignedAgentID) },
		"assigned_party_id": func(q db.Task) any { return derefInt(q.AssignedPartyID) },
		"progress":          func(q db.Task) any { return q.Progress },
		"created_at":        func(q db.Task) any { return q.CreatedAt },
		"started_at":        func(q db.Task) any { return derefTime(q.StartedAt) },
		"completed_at":      func(q db.Task) any { return derefTime(q.CompletedAt) },
	}

	// A marker per position, unique across the row and unique per type, so a value landing in
	// the wrong field of the same type cannot be masked by the two values coinciding.
	markerInt := func(position int) int { return 1000 + position }
	markerString := func(position int) string { return "marker-" + strconv.Itoa(position) }
	markerTime := func(position int) time.Time {
		return time.Date(2026, 1, 1, 0, 0, position, 0, time.UTC)
	}

	var destinations int
	cursor := &stubDailyQuestRowCursor{remaining: 1, scanRow: func(dest ...any) error {
		destinations = len(dest)
		for position, target := range dest {
			switch typed := target.(type) {
			case *int:
				*typed = markerInt(position)
			case *string:
				*typed = markerString(position)
			case **int:
				value := markerInt(position)
				*typed = &value
			case *time.Time:
				*typed = markerTime(position)
			case **time.Time:
				value := markerTime(position)
				*typed = &value
			default:
				t.Fatalf("scan destination %d has unhandled type %T", position, target)
			}
		}
		return nil
	}}

	quests, _, err := scanDailyQuestRows(cursor)
	if err != nil {
		t.Fatalf("scanDailyQuestRows: %v", err)
	}
	if len(quests) != 1 {
		t.Fatalf("read %d quests, want 1", len(quests))
	}
	quest := quests[0]

	if destinations != len(selected) {
		t.Fatalf("activeDailyQuestsQuery selects %d columns (%v) but the scan reads %d destinations -- a positional scan cannot survive that disagreement", len(selected), selected, destinations)
	}

	for position, column := range selected {
		readField, known := fieldHoldingColumn[column]
		if !known {
			t.Errorf("activeDailyQuestsQuery selects %q at position %d, which this test has no db.Task field for -- add it rather than widening the query silently", column, position)
			continue
		}
		got := readField(quest)

		var want any
		switch got.(type) {
		case int:
			want = markerInt(position)
		case string:
			want = markerString(position)
		case time.Time:
			want = markerTime(position)
		default:
			t.Fatalf("column %q reads back as unhandled type %T", column, got)
		}

		if got != want {
			t.Errorf("activeDailyQuestsQuery selects %q at position %d, but position %d of the scan does not land in the %s field (found %v, want %v) -- the SELECT list and the Scan destinations are out of step", column, position, position, column, got, want)
		}
	}
}

// derefInt and derefTime read a nullable field back as a comparable value. A nil here is
// always a failure of the pairing above, never a legitimate NULL: the cursor writes a marker
// into every destination, so any field left nil was not the destination the test thinks it was.
func derefInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func derefTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

// TestTheQuestQueryReadsTheTasksTableTheModelDescribes guards the other half of the pairing:
// the SELECT list names columns db.Task actually has a field for. selectedColumns fails the
// test on anything that is not a bare column name, so this also pins the query to a shape the
// parser above can read.
func TestTheQuestQueryReadsTheTasksTableTheModelDescribes(t *testing.T) {
	selected := selectedColumns(t, activeDailyQuestsQuery)

	// The json tags on db.Task are the model's own names for these columns.
	modelled := []string{
		"id", "name", "description", "difficulty", "xp_reward", "gold_reward", "status",
		"assigned_agent_id", "assigned_party_id", "progress", "created_at", "started_at",
		"completed_at",
	}
	for _, column := range selected {
		if !contains(modelled, column) {
			t.Errorf("activeDailyQuestsQuery selects %q, which db.Task has no field for", column)
		}
	}
	if len(selected) == 0 {
		t.Fatal("parsed no columns out of activeDailyQuestsQuery")
	}
}

// TestANilDatabaseReportsNoLostRows is the boundary arm on the guard clause. A nil database
// has read nothing, so it has lost nothing, and reporting a lost row there would be an
// invented number.
func TestANilDatabaseReportsNoLostRows(t *testing.T) {
	manager := NewDailyQuestManager(nil)

	quests, unreadableRows, err := manager.GetActiveDailyQuests()
	if err != nil {
		t.Fatalf("GetActiveDailyQuests with a nil database: %v", err)
	}
	if quests == nil {
		t.Error("returned a nil slice; the guard clause returns an empty one so the caller can encode it as []")
	}
	if len(quests) != 0 {
		t.Errorf("returned %d quests from a nil database", len(quests))
	}
	if unreadableRows != 0 {
		t.Errorf("reported %d unreadable rows from a nil database, want 0", unreadableRows)
	}
}

// TestTheCountAndTheListAskTheSameQuestion is the cross-endpoint pin, and it is the reason
// the unreadable-row count at this site is more than hygiene.
//
// GetQuestStats publishes active_daily_quests as a SQL COUNT(*); GetActiveDailyQuests returns
// the rows. Both are answers about the same population, so their predicates are one clause
// written twice. Measured on this branch, the two WHERE clauses are identical
// character-for-character -- which means a row that will not scan is counted by
// /api/daily-quests/stats and absent from /api/daily-quests, and before the repair nothing on
// either endpoint said the two could disagree.
//
// The comparison is normalised on whitespace only. Reordering the AND terms would make the
// clauses equivalent and this test red, which is the right way round: it is cheap to keep them
// written the same way, and a reader comparing two endpoints should not have to prove
// equivalence by hand.
func TestTheCountAndTheListAskTheSameQuestion(t *testing.T) {
	listPredicate := wherePredicate(t, activeDailyQuestsQuery)
	countPredicate := wherePredicate(t, activeDailyQuestCountQuery)

	if listPredicate != countPredicate {
		t.Errorf("the list and the count select over different predicates, so the two endpoints can disagree with nothing saying so:\n  list  %s\n  count %s", listPredicate, countPredicate)
	}
	if listPredicate == "" {
		t.Fatal("parsed an empty predicate out of activeDailyQuestsQuery; the parser below is wrong, not the query")
	}
	if !strings.Contains(countPredicate, "[DAILY]") {
		t.Errorf("the count predicate %q does not filter to daily quests at all", countPredicate)
	}
}

// wherePredicate pulls the WHERE clause out of a statement, normalised on whitespace. It is
// deliberately narrow and fails loudly on a shape it cannot read, rather than returning an
// empty string that would make the comparison above pass for the wrong reason.
func wherePredicate(t *testing.T, query string) string {
	t.Helper()
	match := regexp.MustCompile(`(?is)\bWHERE\b(.*?)(?:\bORDER\s+BY\b|$)`).FindStringSubmatch(query)
	if match == nil {
		t.Fatalf("could not find a WHERE clause in %q", query)
	}
	predicate := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(match[1], " "))
	if predicate == "" {
		t.Fatalf("parsed an empty WHERE clause out of %q", query)
	}
	return predicate
}
