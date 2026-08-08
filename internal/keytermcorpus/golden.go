package keytermcorpus

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// updateGoldens regenerates the golden files instead of asserting against them.
//
// Run it for one package at a time, because a package that does not import this
// one does not define the flag:
//
//	go test ./internal/inber      -run TestExtractKeyTermsForNamingMatchesItsGolden -update-key-term-goldens
//	go test ./internal/questgiver -run TestExtractKeyTermsMatchesItsGolden          -update-key-term-goldens
//
// Regenerate both together and read the resulting diff of the two files. That
// diff is the divergence table, and a change to it is the only warning anyone
// gets that unifying the two extractors changed an answer.
var updateGoldens = flag.Bool("update-key-term-goldens", false,
	"rewrite the key-term characterisation golden files from the current behaviour")

// CheckGolden runs every corpus input through extract and compares the answers
// to the golden file at goldenPath, or rewrites that file when
// -update-key-term-goldens is set.
//
// extract is the unexported extractor under test, passed in by a test inside
// its own package because neither function is exported.
//
// The comparison is line by line and reports every difference rather than
// stopping at the first, because the point of a characterisation test is to
// show the whole shape of a behaviour change at once.
func CheckGolden(t *testing.T, goldenPath string, extract func(string) []string) {
	t.Helper()

	lines := make([]string, 0, len(Inputs))
	for _, input := range Inputs {
		encoded, err := json.Marshal(extract(input.Text))
		if err != nil {
			t.Fatalf("encoding the answer for %q: %v", input.Name, err)
		}
		lines = append(lines, input.Name+"\t"+string(encoded))
	}
	got := strings.Join(lines, "\n") + "\n"

	if *updateGoldens {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("creating the golden directory: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("writing %s: %v", goldenPath, err)
		}
		t.Logf("rewrote %s", goldenPath)
		return
	}

	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("reading %s: %v (run with -update-key-term-goldens to create it)", goldenPath, err)
	}

	gotLines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	wantLines := strings.Split(strings.TrimSuffix(string(wantBytes), "\n"), "\n")

	if len(gotLines) != len(wantLines) {
		t.Fatalf("%s has %d entries, the corpus has %d — regenerate it with -update-key-term-goldens",
			goldenPath, len(wantLines), len(gotLines))
	}
	for i := range gotLines {
		if gotLines[i] != wantLines[i] {
			t.Errorf("entry %d changed answer:\n  input:  %q\n  golden: %s\n  now:    %s\n  reaches: %s",
				i, Inputs[i].Text, wantLines[i], gotLines[i], Inputs[i].Reaches)
		}
	}
}

// ReadGolden returns the answers recorded in a golden file, keyed by corpus
// entry name and in corpus order.
//
// It exists so the divergence test can read both files without either extractor
// being importable — the two functions are unexported and live in packages that
// cannot import each other.
func ReadGolden(t *testing.T, goldenPath string) []string {
	t.Helper()

	contents, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("reading %s: %v", goldenPath, err)
	}
	lines := strings.Split(strings.TrimSuffix(string(contents), "\n"), "\n")
	if len(lines) != len(Inputs) {
		t.Fatalf("%s has %d entries, the corpus has %d", goldenPath, len(lines), len(Inputs))
	}

	answers := make([]string, len(lines))
	for i, line := range lines {
		name, answer, found := strings.Cut(line, "\t")
		if !found {
			t.Fatalf("%s line %d is not <name>\\t<answer>: %q", goldenPath, i+1, line)
		}
		if name != Inputs[i].Name {
			t.Fatalf("%s line %d names %q, the corpus names %q — the files are no longer comparable line by line",
				goldenPath, i+1, name, Inputs[i].Name)
		}
		answers[i] = answer
	}
	return answers
}
