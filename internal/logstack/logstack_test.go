package logstack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A minimal session file in the exact shape parseSessionFile accepts: one
// "session" event to supply the id and start time, and one "message" event
// whose content is a []ContentBlock. A fixture in any other shape parses to
// nothing, and a test built on it would pass whether or not the code leaked —
// the positive control below exists precisely to prove this fixture is read.
func writeSession(t *testing.T, dir, marker string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	body := `{"type":"session","id":"sid-` + marker + `","timestamp":"2026-08-17T00:00:00Z"}
{"type":"message","id":"m1","timestamp":"2026-08-17T00:00:01Z","message":{"role":"user","content":[{"type":"text","text":"` + marker + `"}]}}
`
	if err := os.WriteFile(filepath.Join(dir, "s.jsonl"), []byte(body), 0o644); err != nil {
		t.Fatalf("write session: %v", err)
	}
}

// TestGetAgentConversationsReadsALegitimateAgent is the known-positive control.
// It must pass on the fixed code and on main alike: if it ever fails, the two
// traversal tests below are meaningless, because a client that reads nothing
// cannot leak anything and would report "contained" for free.
func TestGetAgentConversationsReadsALegitimateAgent(t *testing.T) {
	agents := t.TempDir()
	writeSession(t, filepath.Join(agents, "goodagent", "sessions"), "INSIDE-OK")
	lc := &LogstackClient{AgentsDir: agents}

	convs, err := lc.GetAgentConversations("goodagent", 20)
	if err != nil {
		t.Fatalf("legitimate agent: unexpected error: %v", err)
	}
	if len(convs) != 1 {
		t.Fatalf("legitimate agent: got %d conversations, want 1", len(convs))
	}
	if got := convs[0].Messages[0].Content; got != "INSIDE-OK" {
		t.Fatalf("legitimate agent: content = %q, want INSIDE-OK", got)
	}
}

// TestGetAgentConversationsRefusesTraversal is the defect. The agent id arrives
// from the `agent` query parameter and is joined onto AgentsDir; a "../"-laden
// value reads outside it. The sentinel is a valid session file one level ABOVE
// AgentsDir, so a traversal that resolves would read it back.
//
// Two assertions, and both matter. The error assertion pins that the guard
// fired; the no-leak assertion pins that it fired for the right reason. Without
// the second, deleting the guard and having the read fail for some unrelated
// reason would still look "refused".
func TestGetAgentConversationsRefusesTraversal(t *testing.T) {
	parent := t.TempDir()
	agents := filepath.Join(parent, "agents")
	if err := os.MkdirAll(agents, 0o755); err != nil {
		t.Fatal(err)
	}
	// Sentinel sits in parent/secret, a sibling of agents — reachable only by
	// escaping AgentsDir.
	writeSession(t, filepath.Join(parent, "secret", "sessions"), "SENTINEL-LEAK")
	lc := &LogstackClient{AgentsDir: agents}

	traversal := filepath.Join("..", "secret")
	convs, err := lc.GetAgentConversations(traversal, 20)
	if err == nil {
		t.Fatalf("traversal %q: expected an error, got nil", traversal)
	}
	if len(convs) != 0 {
		t.Fatalf("traversal %q: leaked %d conversations", traversal, len(convs))
	}
	for _, c := range convs {
		for _, m := range c.Messages {
			if strings.Contains(m.Content, "SENTINEL-LEAK") {
				t.Fatalf("traversal %q: sentinel content leaked", traversal)
			}
		}
	}
}

// TestIsCleanAgentID pins the predicate directly, so a drift in any one branch
// is caught by a named case rather than only where the value happens to be
// exercised end to end.
func TestIsCleanAgentID(t *testing.T) {
	clean := []string{"agent1", "openclaw-main", "a.b.c", "danu"}
	dirty := []string{"", ".", "..", "../secret", "a/b", `a\b`, "sessions/../..",
		"x\x00y", "../../etc"}
	for _, id := range clean {
		if !isCleanAgentID(id) {
			t.Errorf("isCleanAgentID(%q) = false, want true", id)
		}
	}
	for _, id := range dirty {
		if isCleanAgentID(id) {
			t.Errorf("isCleanAgentID(%q) = true, want false", id)
		}
	}
}
