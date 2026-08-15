package inber

import (
	"testing"

	"github.com/kayushkin/inber-party/internal/keytermcorpus"
)

// TestExtractKeyTermsForNamingMatchesItsGolden pins what
// extractKeyTermsForNaming answers today for every input in the shared corpus.
//
// It asserts nothing about whether those answers are right. questgiver has a
// second copy of this function that answers differently, and deciding which of
// the two behaviours is intended needs a person. This test exists so that
// decision can be carried out without guessing: the same corpus is run through
// both functions, and diffing the two golden files shows exactly which inputs
// the two disagree on. See internal/keytermcorpus for the whole argument.
//
// A failure here is not automatically a defect. It means an answer changed, and
// the failure message says which input, what it used to answer, and which
// difference between the two extractors that input was chosen to reach.
func TestExtractKeyTermsForNamingMatchesItsGolden(t *testing.T) {
	keytermcorpus.CheckGolden(t, "testdata/extract_key_terms_for_naming.golden", extractKeyTermsForNaming)
}

// TestExtractKeyTermsForNamingNeverReturnsNothing pins the property that
// separates this extractor from questgiver's most sharply, and that a golden
// file states only by accident.
//
// The fallback chain at the end of the function guarantees at least one term
// for any input at all, including the empty string. questgiver's copy has no
// such chain and returns nil. A caller written against one and pointed at the
// other inherits an empty quest name, so unifying them has to answer this
// deliberately rather than by whichever body survives the merge.
func TestExtractKeyTermsForNamingNeverReturnsNothing(t *testing.T) {
	for _, input := range keytermcorpus.Inputs {
		if got := extractKeyTermsForNaming(input.Text); len(got) == 0 {
			t.Errorf("%s: extractKeyTermsForNaming(%q) returned no terms, but its fallback chain promises at least one",
				input.Name, input.Text)
		}
	}
}

// TestExtractKeyTermsForNamingReturnsAtMostTwoTerms pins the other half of that
// pair: this extractor stops at 2 terms and questgiver's stops at 3. Both
// limits are written as a bare number inside a loop rather than as a named
// constant, so nothing but a test states them.
func TestExtractKeyTermsForNamingReturnsAtMostTwoTerms(t *testing.T) {
	for _, input := range keytermcorpus.Inputs {
		if got := extractKeyTermsForNaming(input.Text); len(got) > 2 {
			t.Errorf("%s: extractKeyTermsForNaming(%q) returned %d terms %q, want at most 2",
				input.Name, input.Text, len(got), got)
		}
	}
}
