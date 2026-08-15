package questgiver

import (
	"strings"
	"testing"
)

// The two call sites in this file are the only ones in either repo that a word
// carrying an apostrophe can actually reach. Everywhere else the input is
// either a fixed vocabulary or guarded by isAlpha, which rejects an apostrophe
// before it arrives. So these are the tests that cover the defect end to end
// rather than at the helper.

// TestExtractKeyTermsKeepsAnApostropheInsideAWord covers questgiver.go:680.
// extractKeyTerms splits on whitespace and filters by length and a stop-word
// list only — there is no isAlpha guard and no punctuation trimming, so a word
// like "don't" reaches the capitaliser intact.
func TestExtractKeyTermsKeepsAnApostropheInsideAWord(t *testing.T) {
	got := extractKeyTerms("don't break the parser")

	var found string
	for _, term := range got {
		if strings.HasPrefix(strings.ToLower(term), "don") {
			found = term
		}
	}
	if found == "" {
		t.Fatalf("extractKeyTerms did not return the apostrophe word at all: %q", got)
	}
	if found == "Don'T" {
		t.Errorf(`extractKeyTerms produced %q — strings.Title capitalised the letter after the apostrophe`, found)
	}
	if found != "Don't" {
		t.Errorf("extractKeyTerms returned %q, want %q", found, "Don't")
	}
}

// TestGenericEpicNameKeepsAnApostropheInsideAWord covers questgiver.go:662,
// which lower-cases the term extracted above and capitalises it again. The
// defect survives that round trip, so fixing only one of the two sites would
// leave the quest name wrong.
func TestGenericEpicNameKeepsAnApostropheInsideAWord(t *testing.T) {
	qng := NewQuestNameGenerator()
	analysis := &TaskAnalysis{TaskType: "debugging", Complexity: 3}

	name := qng.generateGenericEpicName(analysis, "don't break the parser")

	if strings.Contains(name, "'T") {
		t.Errorf("quest name %q contains \"'T\" — the apostrophe was treated as a word break", name)
	}
	if !strings.Contains(name, "Don't") {
		t.Errorf("quest name %q does not contain %q", name, "Don't")
	}
}

// TestWordsWithoutAnApostropheAreUnchanged is the guard against over-correcting.
// The replacement was chosen to preserve every other rendering decision
// strings.Title made, so an ordinary word must come back exactly as before.
func TestWordsWithoutAnApostropheAreUnchanged(t *testing.T) {
	got := extractKeyTerms("rebuild the payment dashboard")
	want := []string{"Rebuild", "Payment", "Dashboard"}

	if len(got) != len(want) {
		t.Fatalf("extractKeyTerms returned %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("term %d = %q, want %q", i, got[i], want[i])
		}
	}
}
