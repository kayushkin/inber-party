package questgiver

import (
	"testing"

	"github.com/kayushkin/inber-party/internal/keytermcorpus"
)

// TestExtractKeyTermsMatchesItsGolden pins what extractKeyTerms answers today
// for every input in the shared corpus.
//
// It is the other half of a pair: internal/inber has a second copy of this
// function, extractKeyTermsForNaming, and the two disagree on every filter they
// carry. Both are run over the same corpus and diffing the two golden files is
// the divergence table. See internal/keytermcorpus for the whole argument.
//
// A failure here is not automatically a defect. It means an answer changed, and
// the failure message says which input, what it used to answer, and which
// difference between the two extractors that input was chosen to reach.
func TestExtractKeyTermsMatchesItsGolden(t *testing.T) {
	keytermcorpus.CheckGolden(t, "testdata/extract_key_terms.golden", extractKeyTerms)
}

// TestExtractKeyTermsReturnsAtMostThreeTerms pins this extractor's limit
// against inber's 2. The number is written inline in the loop, so nothing but a
// test states it.
func TestExtractKeyTermsReturnsAtMostThreeTerms(t *testing.T) {
	for _, input := range keytermcorpus.Inputs {
		if got := extractKeyTerms(input.Text); len(got) > 3 {
			t.Errorf("%s: extractKeyTerms(%q) returned %d terms %q, want at most 3",
				input.Name, input.Text, len(got), got)
		}
	}
}

// TestExtractKeyTermsAppliesNoPunctuationTrimOrAlphabeticGuard pins the absence
// of the two guards inber has, which is the difference that decided where the
// apostrophe defect was reachable.
//
// It is written as an assertion about what comes back rather than about the
// source, so removing either guard from inber — or adding one here — reddens it.
// The corpus entry carries a word ending in a full stop and a word containing a
// digit; inber rejects both and this extractor returns them verbatim apart from
// capitalisation.
func TestExtractKeyTermsAppliesNoPunctuationTrimOrAlphabeticGuard(t *testing.T) {
	got := extractKeyTerms("update the oauth2 handler.")

	var sawDigitWord, sawTrailingFullStop bool
	for _, term := range got {
		if term == "Oauth2" {
			sawDigitWord = true
		}
		if term == "Handler." {
			sawTrailingFullStop = true
		}
	}
	if !sawDigitWord {
		t.Errorf("extractKeyTerms(%q) = %q, want it to contain %q — this extractor has no isAlpha guard",
			"update the oauth2 handler.", got, "Oauth2")
	}
	if !sawTrailingFullStop {
		t.Errorf("extractKeyTerms(%q) = %q, want it to contain %q — this extractor trims no punctuation",
			"update the oauth2 handler.", got, "Handler.")
	}
}
