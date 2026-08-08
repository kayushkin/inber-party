// Package keytermcorpus holds the one shared list of inputs that the
// characterisation tests for both key-term extractors run against.
//
// This repository has two functions that do the same job:
//
//   - internal/inber.extractKeyTermsForNaming
//   - internal/questgiver.extractKeyTerms
//
// One was copied from the other and then edited on one side only. They now
// share 48 stop words and disagree on every filter around them: punctuation
// trimming, the isAlpha guard, a priority-word first pass, the direction the
// input is scanned in, how many terms come back, and whether an input that
// matches nothing returns a placeholder or nothing at all.
//
// That divergence is not cosmetic. It decided the blast radius of the
// apostrophe defect that fix/apostrophes-are-not-word-breaks exists for:
// fourteen call sites used strings.Title, and only the two in questgiver could
// ever receive a word containing an apostrophe, because questgiver has no
// isAlpha guard and inber does. One operation reached through two copies of one
// filter produced a defect on one path and not the other.
//
// Merging the two is a decision about which behaviour is intended, and that
// decision is not this package's to make. What this package makes possible is
// taking the decision safely: every input below is run through both functions
// and the answers are written to a golden file per package. Because both golden
// files are keyed by the same names in the same order, diffing them is the
// divergence table:
//
//	diff internal/inber/testdata/extract_key_terms_for_naming.golden \
//	     internal/questgiver/testdata/extract_key_terms.golden
//
// Whoever unifies the two functions can then see, per input, exactly what their
// merge changes rather than guessing.
//
// Nothing outside a _test.go file imports this package, so it is not linked
// into any binary.
package keytermcorpus

// Input is one corpus entry.
//
// Name identifies the entry in both golden files and never changes once
// written, because the two files are compared line by line. Reaches records
// which behavioural difference between the two extractors the entry was chosen
// to exercise, so a later reader can tell a deliberate case from a decorative
// one.
type Input struct {
	Name    string
	Text    string
	Reaches string
}

// Inputs is the shared corpus, ordered so that both golden files can be
// compared line by line. Append to the end; do not reorder or rename, and
// regenerate both goldens together when you do.
//
// The entries are chosen to reach every axis on which the two extractors
// disagree, not to be representative traffic. Several deliberately produce the
// same answer from both functions — those are the controls, and a change that
// makes one of them diverge is as much a signal as a change to a diverging row.
var Inputs = []Input{
	{
		Name:    "priority-word-first-pass",
		Text:    "rebuild the payment dashboard api",
		Reaches: "inber's priority-word pass, which questgiver has no counterpart for: inber answers with the priority words wherever they sit, questgiver with the first three words that survive its stop list.",
	},
	{
		Name:    "term-limit-two-versus-three",
		Text:    "the api dashboard server database frontend",
		Reaches: "the term limit alone. Every word here is a priority word for inber and a non-stop word for questgiver, so the only thing separating the two answers is that inber stops at 2 and questgiver at 3.",
	},
	{
		Name:    "scan-direction-backwards-versus-forwards",
		Text:    "sharpen the rusty gnomish spanner",
		Reaches: "the direction of the main scan. No priority word appears, so inber walks from the end and keeps the last two, while questgiver walks from the start and keeps the first three.",
	},
	{
		Name:    "stop-word-unique-to-inber",
		Text:    "just fix the widget very quickly",
		Reaches: "the 57 stop words inber has and questgiver does not. \"just\", \"fix\" and \"very\" are stop words on one side and ordinary terms on the other.",
	},
	{
		Name:    "stop-word-unique-to-questgiver",
		Text:    "myself and ours and theirs matter",
		Reaches: "the 5 stop words questgiver has and inber does not: myself, ours, yours, hers, theirs. The divergence runs both ways, not just one.",
	},
	{
		Name:    "punctuation-trim-and-isalpha-guard",
		Text:    "update the oauth2 handler.",
		Reaches: "both guards inber has and questgiver lacks at once: the trailing full stop on \"handler.\" is trimmed by one and kept by the other, and \"oauth2\" is rejected by isAlpha on one side and accepted on the other.",
	},
	{
		Name:    "apostrophe-inside-a-word",
		Text:    "don't break the parser",
		Reaches: "the apostrophe defect's actual blast radius. \"don't\" reaches questgiver's capitaliser intact and is rejected by inber's isAlpha guard before it gets near one — which is why the defect was user-visible on one path only.",
	},
	{
		Name:    "apostrophe-possessive-at-word-end",
		Text:    "restore claxon's ruined banner",
		Reaches: "the same asymmetry for a trailing possessive, where strings.Title produced \"Claxon'S\". inber trims the apostrophe as punctuation only at the edges of a word, so \"claxon's\" still fails isAlpha.",
	},
	{
		Name:    "no-qualifying-word-fallback-mystery",
		Text:    "the it is",
		Reaches: "the never-empty guarantee. inber has a fallback chain ending in \"Mystery\" and always returns at least one term; questgiver returns nothing at all.",
	},
	{
		Name:    "no-qualifying-word-fallback-comments",
		Text:    "fix it: comments.js",
		Reaches: "the \"comment\" arm of inber's fallback chain, and questgiver's lack of any. Every word here is a stop word, too short, or rejected by isAlpha on inber's side, and every one of them survives on questgiver's.",
	},
	{
		Name:    "no-qualifying-word-fallback-code",
		Text:    "do it: v2.js",
		Reaches: "the \".js\" arm of the same chain, plus what the capitaliser does to a dotted filename that questgiver never trims.",
	},
	{
		Name:    "empty-input",
		Text:    "",
		Reaches: "the empty input. inber answers \"Mystery\"; questgiver answers with a nil slice.",
	},
	{
		Name:    "single-word-under-the-length-floor",
		Text:    "go",
		Reaches: "the shared \"longer than 2\" floor, and what each does once it has excluded everything.",
	},
	{
		Name:    "letters-beyond-ascii",
		Text:    "résumé the naïve façade",
		Reaches: "non-ASCII letters. isAlpha is unicode.IsLetter rather than an ASCII range, so accented words survive inber's guard, and the capitaliser has to title-case a multi-byte first rune on both sides.",
	},
	{
		Name:    "mixed-case-input-is-lowered-first",
		Text:    "Rebuild The Payment DASHBOARD",
		Reaches: "the shared strings.ToLower at the top of both functions, which throws away the input's own casing before either capitaliser sees it. Each extractor answers this exactly as it answers the all-lowercase spelling of the same sentence — and they still disagree with each other, because lowering is the only step they share here.",
	},
	{
		Name:    "hyphenated-word",
		Text:    "audit the code-introspection layer",
		Reaches: "a hyphen, which the capitaliser treats as a word break on both sides while the two functions disagree about whether the word reaches it at all — isAlpha rejects a hyphen.",
	},
	{
		Name:    "two-plain-words-the-extractors-agree-on",
		Text:    "rusty spanner",
		Reaches: "agreement, which every other entry above was chosen to break. Two alphabetic words, neither a stop word on either side, neither a priority word, no punctuation, and exactly two of them — so the term limits, the guards and the scan directions all have nothing to bite on. A control: it proves the two functions CAN answer alike, which is what makes the divergence of the other entries a measurement rather than a tautology.",
	},
	{
		Name:    "shared-stop-word-the-extractors-agree-on",
		Text:    "the rusty spanner",
		Reaches: "the 48 stop words the two lists share. It is the entry above plus a word both sides drop, so it stays an agreement — and it separates the shared part of the two stop lists from the 62 words only one of them carries.",
	},
	{
		Name:    "words-of-exactly-three-letters",
		Text:    "orb cog axe",
		Reaches: "the length floor at the one width that can tell \"longer than 2\" from \"longer than 3\". Every other entry answers with a word of four letters or more, so before this one the corpus recorded the same eighteen answers whether either floor was 2 or 3, and no golden could see a change to it. Three letters is also the width the two functions' scan directions and term limits are cleanest on: inber walks back from the end and keeps two, questgiver walks forward and keeps three, so one input measures the floor, the direction and both limits at once.",
	},
}
