package api

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// What these tests pin.
//
// GET /api/activity/timeline builds `events` from two row loops, each of which used to skip an
// unscannable row with a bare `continue` and no log, and neither of which checked rows.Err().
// It then published `total` as len(events) -- computed from the very slice the drop shortened.
// Every other list site in this file has a second source that could disagree with it (a
// neighbouring COUNT(*), a stats endpoint); this one has none, so the response was
// self-consistently wrong and nothing in the system could notice. TimelineResponse.UnreadableRows
// is the channel that lets a caller tell a short answer from a complete one.
//
// ⚠️ Two things these tests deliberately do NOT pin, so that nobody reads them as covered:
//
//  1. The `limit` truncation also shortens `events` and therefore `total`, and is not reported
//     anywhere. That is a caller-requested cap the caller already knows about, not a silent
//     drop, so it is out of this repair's scope. TestTheLimitIsNotCountedAsAnUnreadableRow only
//     pins that the two do not get confused with each other.
//  2. A total failure of the AGENT query answers 200 with no agent events and says nothing. It
//     predates this repair, and changing it alters what a published endpoint returns -- a
//     decision, not a repair. It is the same class of bug and is called out in api.go at the
//     site rather than left for a reader to find.

// unscannableAgentStatusFixture inserts one readable and one unreadable row into each of the two
// tables the timeline reads, and returns the number of events the readable rows owe.
//
// The unreadable rows are produced by a NULL `status`, which the PRODUCTION schema permits: both
// tasks.status and agents.status carry a DEFAULT and are not declared NOT NULL, while the handler
// scans each into a non-pointer string. The fixture is db.OpenSQLiteMirrorOfProductionSchema's,
// so that permission is the real schema's and not a hand-written test table's opinion.
func unscannableAgentStatusFixture(t *testing.T, server *Server) {
	t.Helper()
	now := time.Now().UTC()

	// A readable agent: contributes an agent_updated and an agent_active event.
	if _, err := server.DB.Exec(
		`INSERT INTO agents (id, name, title, class, avatar_emoji, status, last_active, updated_at)
		 VALUES (1,'Readable','Bard','support','X','idle',?,?)`, now, now); err != nil {
		t.Fatalf("inserting the readable agent: %v", err)
	}
	// An agent whose status is NULL: the row is returned by the query and cannot be scanned.
	if _, err := server.DB.Exec(
		`INSERT INTO agents (id, name, title, class, avatar_emoji, status, last_active, updated_at)
		 VALUES (2,'Unreadable','Bard','support','X',NULL,?,?)`, now, now); err != nil {
		t.Fatalf("inserting the unreadable agent: %v", err)
	}
	// A readable task: contributes a task_created event.
	if _, err := server.DB.Exec(
		`INSERT INTO tasks (id, name, status, created_at) VALUES (1,'Readable','available',?)`, now); err != nil {
		t.Fatalf("inserting the readable task: %v", err)
	}
	// A task whose status is NULL, carrying all three timestamps -- so the single row it loses
	// costs THREE events, which is what TestUnreadableRowsCountsRowsNotEvents rests on.
	if _, err := server.DB.Exec(
		`INSERT INTO tasks (id, name, status, created_at, started_at, completed_at)
		 VALUES (2,'Unreadable',NULL,?,?,?)`, now, now, now); err != nil {
		t.Fatalf("inserting the unreadable task: %v", err)
	}
}

func timelineResponseFrom(t *testing.T, server *Server, target string) (TimelineResponse, map[string]json.RawMessage) {
	t.Helper()
	request := httptest.NewRequest("GET", target, nil)
	recorder := httptest.NewRecorder()

	server.handleActivityTimeline(recorder, request)

	if recorder.Code != 200 {
		t.Fatalf("GET %s = %d, want 200; body %s", target, recorder.Code, recorder.Body.String())
	}
	var response TimelineResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decoding the timeline response: %v; body %s", err, recorder.Body.String())
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decoding the timeline response as raw fields: %v", err)
	}
	return response, raw
}

// TestTheTimelineReportsARowItCouldNotRead is the end-to-end reproduction, driven through the
// real handler against a real driver and the production schema. Before the repair this same
// fixture answered 200 with total=3 and nothing else -- three events lost and no channel saying
// so. It is the arm that proves the repair reaches the shipped endpoint and not only the
// extracted collectors.
func TestTheTimelineReportsARowItCouldNotRead(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()
	unscannableAgentStatusFixture(t, server)

	response, _ := timelineResponseFrom(t, server, "/api/activity/timeline")

	if response.UnreadableRows != 2 {
		t.Errorf("unreadable_rows = %d after one unreadable task row and one unreadable agent row, want 2", response.UnreadableRows)
	}
	if len(response.Events) != 3 {
		t.Errorf("events = %d, want 3 (one task_created, one agent_updated, one agent_active from the readable rows)", len(response.Events))
	}
}

// TestTheFieldIsWrittenEvenWhenNothingWasLost is the arm the whole design rests on, and the one
// an "only report problems" implementation fails. It reads the RAW JSON rather than the decoded
// struct on purpose: a Go struct always has the field, so decoding into TimelineResponse cannot
// tell a written zero from an omitted one, and `omitempty` would pass this test if it were
// asserted on the struct. Absent means two things -- nothing was lost, and this build does not
// report -- and a caller that cannot separate them is back where the repair started.
func TestTheFieldIsWrittenEvenWhenNothingWasLost(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	_, raw := timelineResponseFrom(t, server, "/api/activity/timeline")

	value, present := raw["unreadable_rows"]
	if !present {
		t.Fatalf("unreadable_rows is absent from a clean response; present-and-zero is a claim, absent is a build that makes no claim (fields: %v)", rawFieldNames(raw))
	}
	if string(value) != "0" {
		t.Errorf("unreadable_rows = %s on a clean read, want 0", value)
	}
}

func rawFieldNames(raw map[string]json.RawMessage) []string {
	names := make([]string, 0, len(raw))
	for name := range raw {
		names = append(names, name)
	}
	return names
}

// TestUnreadableRowsCountsRowsNotEvents pins what the field's name claims. The unreadable task
// row in the fixture carries created_at, started_at AND completed_at, so it would have produced
// three events on its own; the unreadable agent row would have produced two. An implementation
// that counted lost EVENTS would answer 5 here and would be a different, unknowable number --
// the events of a row that could not be read cannot be reconstructed, so counting them would be
// inventing them.
func TestUnreadableRowsCountsRowsNotEvents(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()
	unscannableAgentStatusFixture(t, server)

	response, _ := timelineResponseFrom(t, server, "/api/activity/timeline")

	if response.UnreadableRows != 2 {
		t.Errorf("unreadable_rows = %d, want 2 -- two ROWS were lost; the five events they would have produced are not what this field counts", response.UnreadableRows)
	}
}

// TestTheTotalStillCountsOnlyTheEventsThatWereRead is the do-no-harm arm. The repair records the
// drop and deliberately leaves the count alone; a `total` quietly raised by the shortfall would
// be a promise about events that are not in the array.
func TestTheTotalStillCountsOnlyTheEventsThatWereRead(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()
	unscannableAgentStatusFixture(t, server)

	response, _ := timelineResponseFrom(t, server, "/api/activity/timeline")

	if response.Total != len(response.Events) {
		t.Errorf("total = %d with %d events in the array; total must keep meaning \"events in this response\"", response.Total, len(response.Events))
	}
	if response.Total != 3 {
		t.Errorf("total = %d, want 3 -- the two lost rows must not be added back into it", response.Total)
	}
}

// TestTheLimitIsNotCountedAsAnUnreadableRow separates the two ways this array can be short. The
// limit is a caller-requested cap and is not a drop; folding it into unreadable_rows would make
// the field report a number the caller asked for.
func TestTheLimitIsNotCountedAsAnUnreadableRow(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()
	unscannableAgentStatusFixture(t, server)

	response, _ := timelineResponseFrom(t, server, "/api/activity/timeline?limit=1")

	if len(response.Events) != 1 {
		t.Fatalf("events = %d with limit=1, want 1", len(response.Events))
	}
	if response.UnreadableRows != 2 {
		t.Errorf("unreadable_rows = %d with limit=1, want 2 -- the limit dropped two further events and must not change this count", response.UnreadableRows)
	}
}

// --- the collectors, driven by a stub cursor ---
//
// These arms exist because a real driver gives no reliable way to make rows.Err() non-nil on
// demand. Without them the rows.Err() half of the repair would ship with no test at all, which
// is how it came to be missing.

type stubTimelineCursor struct {
	scanResults []error
	index       int
	iterateErr  error
}

func (c *stubTimelineCursor) Next() bool {
	if c.index >= len(c.scanResults) {
		return false
	}
	c.index++
	return true
}

func (c *stubTimelineCursor) Scan(dest ...interface{}) error {
	return c.scanResults[c.index-1]
}

func (c *stubTimelineCursor) Err() error { return c.iterateErr }

func TestAnUnscannableRowIsCountedNotErased(t *testing.T) {
	scanFailed := errors.New("converting NULL to string is unsupported")
	for _, collector := range []struct {
		name string
		call func(timelineRowCursor, time.Time, time.Time) ([]ActivityEvent, int, error)
	}{
		{"collectTaskTimelineEvents", collectTaskTimelineEvents},
		{"collectAgentTimelineEvents", collectAgentTimelineEvents},
	} {
		t.Run(collector.name, func(t *testing.T) {
			cursor := &stubTimelineCursor{scanResults: []error{scanFailed, scanFailed, scanFailed}}

			events, unreadableRows, err := collector.call(cursor, time.Now().Add(-time.Hour), time.Now())

			if err != nil {
				t.Fatalf("a row that will not scan is not an error for the whole read: %v", err)
			}
			if unreadableRows != 3 {
				t.Errorf("unreadableRows = %d after three rows failed to scan, want 3", unreadableRows)
			}
			if len(events) != 0 {
				t.Errorf("events = %d, want 0 -- no row scanned", len(events))
			}
		})
	}
}

// TestAnIterationFailureIsReturnedNotCounted pins the asymmetry between the two silent drops at
// this site. A row that will not scan is countable, so it is counted. An iteration failure is
// not: the driver stops wherever it failed, so how many rows it cost is unknowable and any count
// would be invented. The honest report is the error.
func TestAnIterationFailureIsReturnedNotCounted(t *testing.T) {
	iterationFailed := errors.New("connection reset by peer")
	for _, collector := range []struct {
		name string
		call func(timelineRowCursor, time.Time, time.Time) ([]ActivityEvent, int, error)
	}{
		{"collectTaskTimelineEvents", collectTaskTimelineEvents},
		{"collectAgentTimelineEvents", collectAgentTimelineEvents},
	} {
		t.Run(collector.name, func(t *testing.T) {
			cursor := &stubTimelineCursor{iterateErr: iterationFailed}

			events, unreadableRows, err := collector.call(cursor, time.Now().Add(-time.Hour), time.Now())

			if err == nil {
				t.Fatalf("an iteration failure returned a nil error; the partial list would go back as if it were complete")
			}
			if !errors.Is(err, iterationFailed) {
				t.Errorf("err = %v, want it to wrap the driver's error so the cause survives", err)
			}
			if unreadableRows != 0 {
				t.Errorf("unreadableRows = %d beside an iteration failure, want 0 -- a count here would be invented", unreadableRows)
			}
			if events != nil {
				t.Errorf("events = %v beside an iteration failure, want nil -- a partial slice returned with an error invites a caller to publish it", events)
			}
		})
	}
}

// TestATimelineReadThatCannotRunDoesNotAnswerTwoHundred is named for what it actually reaches,
// which is NOT the iteration path. Closing the database makes the QUERY fail, so this exercises
// the pre-existing task-query arm; no in-process fixture can make a live SQLite cursor return
// rows.Err() != nil part-way through, so the handler's pass-through of a collector's iteration
// error is a known coverage hole. The scorer's `answer-200-on-an-agent-iteration-failure` arm is
// recorded as UNPINNED and prints that hole on every run rather than leaving it out of the score.
// The collectors' own iteration behaviour is pinned, by the stub-cursor arms above.
func TestATimelineReadThatCannotRunDoesNotAnswerTwoHundred(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	// Closing the database makes the query itself fail rather than the iteration, which is the
	// nearest a real driver gets on demand. The arm that matters is that a read which cannot be
	// completed does not answer 200 with a short array.
	server.DB.Close()

	request := httptest.NewRequest("GET", "/api/activity/timeline", nil)
	recorder := httptest.NewRecorder()

	server.handleActivityTimeline(recorder, request)

	if recorder.Code == 200 {
		t.Errorf("a timeline read that could not run answered 200 with body %s; a short answer that says nothing is the bug this file repairs", strings.TrimSpace(recorder.Body.String()))
	}
}
