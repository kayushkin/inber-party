package inber

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// The defect these tests pin: NewStore skips a database whose file is missing and
// returns a Store anyway, so a read path with no handle had to decide what that
// meant. Every read below used to answer with an empty result, which a caller
// cannot tell apart from a genuinely quiet week. The live quest board was blank
// for that reason and the blankness was reported as success.
//
// Each case asks two things of the answer: that it is an error at all, and that
// the error names the file that was expected. The second half is what makes the
// error worth more than a bare refusal — the failure in production was a path
// still pointing at a database inber had renamed, which only the path reveals.

// storeWithNoDatabases builds a Store the way NewStore does when neither file
// exists: no handles, but both configured paths remembered.
func storeWithNoDatabases(t *testing.T) (store *Store, sessionsPath, gatewayPath string) {
	t.Helper()
	dir := t.TempDir()
	sessionsPath = filepath.Join(dir, "sessions.db")
	gatewayPath = filepath.Join(dir, "gateway", "gateway.db")

	store, err := NewStore(sessionsPath, gatewayPath, "")
	if err != nil {
		t.Fatalf("NewStore with absent files: %v", err)
	}
	if store.sessionsDB != nil || store.gatewayDB != nil {
		t.Fatalf("fixture is not exercising the case under test: a handle was opened")
	}
	return store, sessionsPath, gatewayPath
}

func TestReadsRefuseWhenTheDatabaseWasNeverOpened(t *testing.T) {
	store, sessionsPath, gatewayPath := storeWithNoDatabases(t)
	defer store.Close()

	cases := []struct {
		name     string
		wantPath string
		call     func() (any, error)
	}{
		{"GetQuests", gatewayPath, func() (any, error) { return store.GetQuests(10) }},
		{"GetQuestHistory", gatewayPath, func() (any, error) { return store.GetQuestHistory("brigid", 10) }},
		{"GetSessionReplay", gatewayPath, func() (any, error) { return store.GetSessionReplay("some-session") }},
		{"GetAgentJournal", gatewayPath, func() (any, error) { return store.GetAgentJournal("brigid", "2026-08-22") }},
		{"GetConversations", sessionsPath, func() (any, error) { return store.GetConversations(10) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.call()
			if err == nil {
				t.Fatalf("%s returned %#v and no error; an unopened database must not read as an empty world", tc.name, got)
			}
			if !strings.Contains(err.Error(), tc.wantPath) {
				t.Errorf("%s error does not name the file it expected.\n got: %v\nwant it to contain: %s", tc.name, err, tc.wantPath)
			}
			if !strings.Contains(err.Error(), "not open") {
				t.Errorf("%s error does not say the database is not open: %v", tc.name, err)
			}
		})
	}
}

// The journal is called out separately because its old answer was not merely
// empty, it was a fabrication: a narrative asserting the agent had rested.
func TestTheJournalDoesNotNarrateAnUnopenedDatabaseAsARestfulDay(t *testing.T) {
	store, _, _ := storeWithNoDatabases(t)
	defer store.Close()

	journal, err := store.GetAgentJournal("brigid", "2026-08-22")
	if err == nil {
		t.Fatalf("GetAgentJournal invented a journal from no database: %+v", journal)
	}
	if journal != nil {
		t.Errorf("GetAgentJournal returned both an error and a journal: %+v", journal)
	}
	if strings.Contains(err.Error(), "archives are silent") {
		t.Errorf("the fabricated narrative is still the answer, only wearing an error: %v", err)
	}
}

// A read whose database IS open must not be caught by the new guard. Without this
// the suite would pass just as well if every read refused unconditionally.
func TestAnOpenDatabaseStillAnswers(t *testing.T) {
	gatewayPath := createTestGatewayDB(t)
	store, err := NewStore("", gatewayPath, "")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	quests, err := store.GetQuests(10)
	if err != nil {
		t.Fatalf("GetQuests against an open gateway database: %v", err)
	}
	if len(quests) == 0 {
		t.Fatal("fixture supplied no quests, so this control cannot tell a working read from a refusing one")
	}

	if _, err := store.GetQuestHistory("brigid", 10); err != nil {
		t.Errorf("GetQuestHistory against an open gateway database: %v", err)
	}
}

// The sessions and gateway databases are refused independently: opening one must
// not make a read of the other look available.
func TestOneOpenDatabaseDoesNotVouchForTheOther(t *testing.T) {
	gatewayPath := createTestGatewayDB(t)
	sessionsPath := filepath.Join(t.TempDir(), "sessions.db")

	store, err := NewStore(sessionsPath, gatewayPath, "")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if store.gatewayDB == nil {
		t.Fatal("fixture failed: the gateway database should be open")
	}
	if store.sessionsDB != nil {
		t.Fatal("fixture failed: the sessions database should be absent")
	}

	if _, err := store.GetQuests(10); err != nil {
		t.Errorf("the open gateway database should still answer: %v", err)
	}
	if _, err := store.GetConversations(10); err == nil {
		t.Error("GetConversations answered from a sessions database that was never opened")
	}
}

// errDatabaseNotOpen distinguishes "the file was not there" from "no path was
// configured at all". Both are refusals; only the first has a path to report.
func TestTheRefusalSaysWhichOfTheTwoWaysItFailed(t *testing.T) {
	withPath := errDatabaseNotOpen("gateway", "/home/someone/.inber/gateway/gateway.db")
	if !strings.Contains(withPath.Error(), "/home/someone/.inber/gateway/gateway.db") {
		t.Errorf("a configured path must appear in the error: %v", withPath)
	}

	withoutPath := errDatabaseNotOpen("gateway", "")
	if !strings.Contains(withoutPath.Error(), "no path configured") {
		t.Errorf("an unconfigured database must say so rather than quote an empty path: %v", withoutPath)
	}
	if strings.Contains(withoutPath.Error(), `""`) {
		t.Errorf("an unconfigured database quoted an empty path: %v", withoutPath)
	}

	if errors.Is(withPath, withoutPath) {
		t.Error("the two refusals compare equal, so a caller cannot tell them apart")
	}
}
