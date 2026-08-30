package api

import (
	"net/http/httptest"
	"testing"
)

// What these tests pin.
//
// GET /api/daily-quests answers a bare JSON array. When GetActiveDailyQuests could not read a
// row, the array is short and there is nothing in the response that says so -- and the
// neighbouring /api/daily-quests/stats publishes active_daily_quests as a SQL COUNT(*) over
// the same predicate, so the two endpoints disagree. setUnreadableRowsHeader is the channel
// that lets a caller tell a short list from a complete one without changing the body's shape.

// TestTheHeaderIsWrittenEvenWhenNothingWasLost is the arm the whole design rests on, and it is
// the one an "only report problems" implementation fails. An absent header would mean two
// different things -- nothing was lost, and this build does not report -- and a caller that
// cannot separate those is back where the repair started.
func TestTheHeaderIsWrittenEvenWhenNothingWasLost(t *testing.T) {
	recorder := httptest.NewRecorder()

	setUnreadableRowsHeader(recorder, 0)

	values, ok := recorder.Result().Header[unreadableRowsHeader]
	if !ok {
		t.Fatalf("%s is absent after a clean read; present-and-zero is a claim, absent is a build that makes no claim", unreadableRowsHeader)
	}
	if len(values) != 1 || values[0] != "0" {
		t.Errorf("%s = %v, want exactly [\"0\"]", unreadableRowsHeader, values)
	}
}

// TestTheHeaderCarriesTheCountNotAFlag is the discriminating positive. A boolean "some rows
// were lost" would pass the arm above and lose the only number the caller can compare against
// the stats endpoint's COUNT(*).
func TestTheHeaderCarriesTheCountNotAFlag(t *testing.T) {
	for _, unreadableRows := range []int{1, 2, 17} {
		recorder := httptest.NewRecorder()

		setUnreadableRowsHeader(recorder, unreadableRows)

		got := recorder.Result().Header.Get(unreadableRowsHeader)
		want := map[int]string{1: "1", 2: "2", 17: "17"}[unreadableRows]
		if got != want {
			t.Errorf("%s = %q after losing %d rows, want %q -- the count is what makes the shortfall comparable with active_daily_quests", unreadableRowsHeader, got, unreadableRows, want)
		}
	}
}

// TestTheHeaderDistinguishesEveryCount is the arm that catches a constant. An implementation
// that always wrote "0", or always wrote "1", passes one of the two tests above; nothing but
// requiring the values to differ from each other catches both.
func TestTheHeaderDistinguishesEveryCount(t *testing.T) {
	seen := map[string]int{}
	for _, unreadableRows := range []int{0, 1, 2, 3} {
		recorder := httptest.NewRecorder()
		setUnreadableRowsHeader(recorder, unreadableRows)
		seen[recorder.Result().Header.Get(unreadableRowsHeader)]++
	}
	if len(seen) != 4 {
		t.Errorf("four different row counts produced %d distinct header values (%v), want 4 -- a header that cannot separate them reports nothing", len(seen), seen)
	}
}

// TestTheListEndpointStillAnswersAnArray is the wire-shape control. The repair deliberately
// did NOT turn the body into an object, because that is a decision on a published endpoint
// rather than a repair. If somebody later makes that change, this is the test that says so
// out loud instead of letting a caller find out.
func TestTheListEndpointStillAnswersAnArray(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	request := httptest.NewRequest("GET", "/api/daily-quests", nil)
	recorder := httptest.NewRecorder()

	server.handleDailyQuests(recorder, request)

	body := recorder.Body.String()
	if len(body) == 0 || (body[0] != '[' && body[0] != 'n') {
		t.Errorf("GET /api/daily-quests answered %q; this endpoint's body is a JSON array and changing that is a wire change, not a repair", body)
	}
}
