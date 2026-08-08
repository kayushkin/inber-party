package keytermcorpus_test

import (
	"sort"
	"testing"

	"github.com/kayushkin/inber-party/internal/keytermcorpus"
)

const (
	inberGolden      = "../inber/testdata/extract_key_terms_for_naming.golden"
	questgiverGolden = "../questgiver/testdata/extract_key_terms.golden"
)

// divergentInputs lists, by corpus entry name, every input the two extractors
// answer differently today.
//
// It is the measurement this whole package exists to take, written down so that
// it can change loudly. Neither extractor is importable from the other's
// package and both are unexported, so this is the only place in the repository
// where the two behaviours can be compared at all.
//
// Adding a name here means the two functions have drifted further apart.
// Removing one means a change brought them closer — which is the direction
// anyone unifying them is going, and the reason this is a list rather than a
// count. A count would say the number moved; the list says which input moved.
var divergentInputs = []string{
	"apostrophe-inside-a-word",
	"apostrophe-possessive-at-word-end",
	"empty-input",
	"hyphenated-word",
	"letters-beyond-ascii",
	"mixed-case-input-is-lowered-first",
	"no-qualifying-word-fallback-code",
	"no-qualifying-word-fallback-comments",
	"no-qualifying-word-fallback-mystery",
	"priority-word-first-pass",
	"punctuation-trim-and-isalpha-guard",
	"scan-direction-backwards-versus-forwards",
	"single-word-under-the-length-floor",
	"stop-word-unique-to-inber",
	"stop-word-unique-to-questgiver",
	"term-limit-two-versus-three",
}

// TestTheTwoExtractorsDivergeOnExactlyTheRecordedInputs compares the two golden
// files line by line and asserts the set of disagreements is the one recorded
// above.
//
// It reads the goldens rather than calling the functions because it cannot call
// them: extractKeyTermsForNaming and extractKeyTerms are unexported, in two
// packages, one of which imports the other. The golden files are the only
// shared surface, which is why both are written from the same corpus in the
// same order.
//
// If this fails after a deliberate change, regenerate both goldens and update
// the list. If it fails without one, two copies of one function have drifted.
func TestTheTwoExtractorsDivergeOnExactlyTheRecordedInputs(t *testing.T) {
	inberAnswers := keytermcorpus.ReadGolden(t, inberGolden)
	questgiverAnswers := keytermcorpus.ReadGolden(t, questgiverGolden)

	var diverged []string
	for i, input := range keytermcorpus.Inputs {
		if inberAnswers[i] != questgiverAnswers[i] {
			diverged = append(diverged, input.Name)
		}
	}
	sort.Strings(diverged)

	want := append([]string(nil), divergentInputs...)
	sort.Strings(want)

	if len(diverged) != len(want) {
		t.Errorf("the two extractors diverge on %d of %d corpus inputs, recorded %d",
			len(diverged), len(keytermcorpus.Inputs), len(want))
	}

	inWant := make(map[string]bool, len(want))
	for _, name := range want {
		inWant[name] = true
	}
	inGot := make(map[string]bool, len(diverged))
	for _, name := range diverged {
		inGot[name] = true
	}

	for i, input := range keytermcorpus.Inputs {
		switch {
		case inGot[input.Name] && !inWant[input.Name]:
			t.Errorf("%s now diverges and is not recorded:\n  input:      %q\n  inber:      %s\n  questgiver: %s\n  reaches:    %s",
				input.Name, input.Text, inberAnswers[i], questgiverAnswers[i], input.Reaches)
		case !inGot[input.Name] && inWant[input.Name]:
			t.Errorf("%s is recorded as diverging but the two now agree on %s:\n  input:   %q\n  reaches: %s",
				input.Name, inberAnswers[i], input.Text, input.Reaches)
		}
	}
}

// TestEveryCorpusEntryIsNamedOnceAndExplained guards the corpus itself. Both
// golden files are compared to it by position, so a duplicate or blank name
// makes the comparison silently meaningless rather than failing.
func TestEveryCorpusEntryIsNamedOnceAndExplained(t *testing.T) {
	seen := make(map[string]int, len(keytermcorpus.Inputs))
	for i, input := range keytermcorpus.Inputs {
		if input.Name == "" {
			t.Errorf("corpus entry %d has no name", i)
			continue
		}
		if input.Reaches == "" {
			t.Errorf("corpus entry %q does not say which difference it reaches", input.Name)
		}
		if first, ok := seen[input.Name]; ok {
			t.Errorf("corpus entries %d and %d are both named %q", first, i, input.Name)
		}
		seen[input.Name] = i
	}
}

// TestEveryRecordedDivergentInputIsInTheCorpus catches the other direction: a
// name in divergentInputs that no corpus entry carries would otherwise sit
// there forever, counted but never compared.
func TestEveryRecordedDivergentInputIsInTheCorpus(t *testing.T) {
	inCorpus := make(map[string]bool, len(keytermcorpus.Inputs))
	for _, input := range keytermcorpus.Inputs {
		inCorpus[input.Name] = true
	}
	for _, name := range divergentInputs {
		if !inCorpus[name] {
			t.Errorf("divergentInputs names %q, which is not a corpus entry", name)
		}
	}
}
