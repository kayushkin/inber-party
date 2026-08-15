#!/usr/bin/env python3
"""Sabotage-score inber.go's pure text and naming ARITHMETIC.

scripts/sabotage-truncation-budgets.py scores the six truncation sites — the
guard that decides whether to cut and the budget handed to the cut. This scores
the numbers *around* them: the complexity ladder that picks a quest-name
pattern, the word-length floors and term limits in the key-term extractor, the
spawn-name floor, and the XP and energy formulas. No mutation here touches a
truncation number, and no case in that scorer touches one of these.

  THE PRIOR COLUMN IS THE WHOLE REPO, NOT THE PACKAGE AND NOT THE FILE.

The 205th pass's rule is that a baseline built as a diff against the one test
file a sweep looks like it extends will report a gap that a SIBLING test already
covers. This package has a live sibling risk: the key-term extractor is pinned
by a golden corpus in extract_key_terms_characterisation_test.go, and four other
packages import internal/inber. So PRIOR renames this sweep's file aside and
runs `go test ./...`. Anything a neighbour already holds reports PRIOR=CAUGHT and
lands in "already covered" rather than in the closed-gaps count. The whole run
costs about a second per case; a flattering number costs a reader much more.

  A DOMINATED MUTATION IS NOT A GAP, AND THREE OF THESE ARE DOMINATED.

analyzeTaskForNaming's clamps run in a fixed order over a ladder that tops out
at 4, so `min(5, complexity+1)` on the urgent line can never cut anything and
`max(1, complexity-1)` can never lift anything. Drifting those constants in the
unreachable direction changes no answer the function can produce, so no test can
catch them — the same shape as the truncation scorer's `maxBytes <= 0 -> < 0`
row. They are scored as declared known-NEGATIVE controls rather than as open
gaps, because filing them as gaps would send the next pass after a test that
cannot exist.

Rules inherited from earlier passes, each of which cost a session to learn:

  - The needle must occur EXACTLY once and the file's bytes must change. Three
    of this file's clamp lines are byte-identical, so every one of them is
    matched with its own `if` line above it; a bare needle would silently
    mutate the first of the three.
  - A compile error is not a red. It is a separate verdict.
  - A panic in a source frame is detection; a panic in a _test.go frame is the
    fixture falling over and is not.
  - Two controls, a known-positive and a known-negative, so a harness that
    reports CAUGHT for everything cannot look perfect.
  - Restores from git, so the tree must be committed before running.

Run:  python3 scripts/sabotage-naming-arithmetic.py
      python3 scripts/sabotage-naming-arithmetic.py --self-test
"""

import os
import pathlib
import re
import signal
import subprocess
import sys

REPO = pathlib.Path(__file__).resolve().parent.parent

INBER = "internal/inber/inber.go"

TRACKED = [INBER]

# The whole repo, for the reason in the docstring: four packages import
# internal/inber, and a mutation any of them catches is not a gap this sweep
# closed.
PACKAGES = ["./..."]

# The file this sweep adds. Renamed aside to build the PRIOR column.
MINE = ["internal/inber/naming_arithmetic_boundaries_test.go"]

# Reach guards, by message. This sweep's tests call their targets directly and
# need no fixture to reach them, so it contributes none — the entry below is
# inherited from the truncation tests, which share these packages.
GUARD_MARKERS = (
    'input no longer reaches the truncating branch',
    "no longer reaches generateQuestName's truncating fallback",
)

_FAIL_LINE = re.compile(r"^\s*(\S+_test\.go):(\d+): (.*)$", re.M)
# A stack frame naming a .go file. The absolute path is captured because the Go
# runtime's own frames (runtime/panic.go) would pass a bare-filename filter.
_FRAME = re.compile(r"^\s+(/\S+\.go):(\d+)", re.M)


def verdict_for(output, returncode):
    """CAUGHT / CAUGHT (guard) / UNNOTICED / COMPILE ERROR, from one run."""
    if "build failed" in output or "[build failed]" in output or "syntax error" in output:
        return "COMPILE ERROR"
    if returncode == 0:
        return "UNNOTICED"
    messages = [m for _, _, m in _FAIL_LINE.findall(output)]
    guard = [m for m in messages if any(g in m for g in GUARD_MARKERS)]
    real = [m for m in messages if m not in guard]
    if real:
        return "CAUGHT"
    if "panic:" in output:
        frames = [f for f, _ in _FRAME.findall(output.split("panic:", 1)[1])
                  if f.startswith(str(REPO) + "/")]
        source = next((f for f in frames if not f.endswith("_test.go")), None)
        if source:
            return "CAUGHT (panic in %s)" % pathlib.Path(source).name
        return "CAUGHT (fixture panicked)"
    if guard:
        return "CAUGHT (guard)"
    return "CAUGHT (no message)"


def counts_as_coverage(verdict):
    """Whether a verdict means an assertion actually looked at the value."""
    return verdict == "CAUGHT" or verdict.startswith("CAUGHT (panic in ")


def self_test():
    """No case table can score its own scorer, so drive every verdict."""
    probes = [
        ("--- FAIL: T\n    x_test.go:9: got 1 want 2\n", 1, "CAUGHT"),
        ("--- FAIL: T\n    x_test.go:9: input no longer reaches the truncating branch\n", 1,
         "CAUGHT (guard)"),
        ("--- FAIL: T\n    x_test.go:9: input no longer reaches the truncating branch\n"
         "    x_test.go:12: got 1 want 2\n", 1, "CAUGHT"),
        ("ok  \tpkg\t0.01s\n", 0, "UNNOTICED"),
        ("# pkg [build failed]\n./x.go:9:2: undefined: y\n", 1, "COMPILE ERROR"),
        ("--- FAIL: T\npanic: boom\n\t/usr/lib/go/src/runtime/panic.go:8\n\t%s/util.go:3\n" % REPO,
         1, "CAUGHT (panic in util.go)"),
        ("--- FAIL: T\npanic: boom\n\t%s/x_test.go:3\n" % REPO, 1, "CAUGHT (fixture panicked)"),
        ("--- FAIL: T\nno test line at all\n", 1, "CAUGHT (no message)"),
    ]
    ok = True
    for output, rc, want in probes:
        got = verdict_for(output, rc)
        if got != want:
            print("SELF-TEST FAIL: verdict_for -> %r, want %r" % (got, want))
            ok = False
    for verdict, want in [("CAUGHT", True), ("CAUGHT (panic in util.go)", True),
                          ("CAUGHT (guard)", False), ("CAUGHT (no message)", False),
                          ("CAUGHT (fixture panicked)", False),
                          ("UNNOTICED", False), ("COMPILE ERROR", False)]:
        if counts_as_coverage(verdict) != want:
            print("SELF-TEST FAIL: counts_as_coverage(%r) != %r" % (verdict, want))
            ok = False
    print("self-test: %s" % ("all verdicts reachable and separated" if ok else "FAILED"))
    return 0 if ok else 1


if "--self-test" in sys.argv:
    sys.exit(self_test())


# The three clamp lines in analyzeTaskForNaming are byte-identical, so each is
# matched together with the `if` line that owns it. Written out once here rather
# than inline, because getting one of them wrong mutates a neighbour's line and
# scores it under this row's name.
URGENT_CLAMP = ('if containsAnyKeyword(lower, []string{"urgent", "critical", "emergency", "asap"}) {\n'
                '\t\tcomplexity = min(5, complexity+1)')
LENGTH_CLAMP = 'if len(text) > 100 {\n\t\tcomplexity = min(5, complexity+1)'
ERROR_CLAMP = 'if status == "error" {\n\t\tcomplexity = min(5, complexity+1)'

# (label, [(file, find, replace)], site)
#
# Every mutation is a DRIFTED VALUE, not a deletion: a number is far likelier to
# drift by one than a whole check is to disappear, and a drift keeps every
# identifier live so the case reports a score instead of a compile error.
CASES = [
    # -- xpForTokens: the /100 divisor --------------------------------------
    # The prior suite checks 0->0, 100->1, 1500->15. Because 1500/99 == 15 and
    # 100/99 == 1, the DOWN mutation survives all three and the UP one does not.
    # Scoring both is what shows the pin was one-sided.
    ("the XP divisor drifts down, so 99 tokens already earn a whole XP",
     [(INBER, "return tokens / 100", "return tokens / 99")],
     "xpForTokens divisor"),

    ("the XP divisor drifts up, so 100 tokens earn nothing",
     [(INBER, "return tokens / 100", "return tokens / 101")],
     "xpForTokens divisor"),

    # -- energyFromActivity: four numbers a range check cannot separate -----
    ("the never-seen answer drifts, so an unseen agent is one short of full",
     [(INBER, "if lastActive == nil {\n\t\treturn 100",
       "if lastActive == nil {\n\t\treturn 99")],
     "energy nil answer"),

    ("the energy cap drifts down by one",
     [(INBER, "math.Min(100, 20+hoursSince*10)", "math.Min(99, 20+hoursSince*10)")],
     "energy cap"),

    ("the energy floor drifts up, so a just-active agent starts at 21",
     [(INBER, "math.Min(100, 20+hoursSince*10)", "math.Min(100, 21+hoursSince*10)")],
     "energy floor"),

    ("the recovery rate drifts down to 9 per hour",
     [(INBER, "math.Min(100, 20+hoursSince*10)", "math.Min(100, 20+hoursSince*9)")],
     "energy rate"),

    # -- analyzeTaskForNaming: the per-type complexity ladder ---------------
    ("the default complexity drifts up",
     [(INBER, '\ttaskType = "general"\n\tcomplexity = 2',
       '\ttaskType = "general"\n\tcomplexity = 3')],
     "complexity ladder"),

    ("debugging's complexity drifts down",
     [(INBER, '\t\ttaskType = "debugging"\n\t\tcomplexity = 3',
       '\t\ttaskType = "debugging"\n\t\tcomplexity = 2')],
     "complexity ladder"),

    ("development's complexity drifts up",
     [(INBER, '\t\ttaskType = "development"\n\t\tcomplexity = 4',
       '\t\ttaskType = "development"\n\t\tcomplexity = 5')],
     "complexity ladder"),

    ("design's complexity drifts down",
     [(INBER, '\t\ttaskType = "design"\n\t\tcomplexity = 4',
       '\t\ttaskType = "design"\n\t\tcomplexity = 3')],
     "complexity ladder"),

    # -- analyzeTaskForNaming: the 100-byte length bump ---------------------
    ("the long-task threshold drifts up, so a 101-byte task is not bumped",
     [(INBER, "if len(text) > 100 {", "if len(text) > 101 {")],
     "length bump"),

    ("the long-task threshold drifts down, so a 100-byte task is bumped",
     [(INBER, "if len(text) > 100 {", "if len(text) > 99 {")],
     "length bump"),

    # -- analyzeTaskForNaming: the clamps that CAN bind ---------------------
    ("the complex-task ceiling rises, so development+complex reaches 6",
     [(INBER, "complexity = min(5, complexity+2)", "complexity = min(6, complexity+2)")],
     "ceiling (complex)"),

    ("the long-task ceiling rises, so an already-clamped 5 is bumped to 6",
     [(INBER, LENGTH_CLAMP, LENGTH_CLAMP.replace("min(5,", "min(6,"))],
     "ceiling (length)"),

    ("the error ceiling rises, so an already-clamped 5 is bumped to 6",
     [(INBER, ERROR_CLAMP, ERROR_CLAMP.replace("min(5,", "min(6,"))],
     "ceiling (error)"),

    # -- the dominated clamps, from the ONE side that is observable ---------
    # The unreachable direction of each is a control below, not a case here.
    ("the urgent ceiling drifts DOWN, which is its one observable direction",
     [(INBER, URGENT_CLAMP, URGENT_CLAMP.replace("min(5,", "min(4,"))],
     "ceiling (urgent)"),

    ("the simple-task floor drifts UP, which is its one observable direction",
     [(INBER, "complexity = max(1, complexity-1)", "complexity = max(2, complexity-1)")],
     "floor (simple)"),

    # -- extractKeyTermsForNaming: two floors and two term limits -----------
    ("the priority-pass word floor rises, so the three-letter priority word is skipped",
     [(INBER, "if priorityWords[cleanWord] && len(cleanWord) > 2 {",
       "if priorityWords[cleanWord] && len(cleanWord) > 3 {")],
     "key terms: priority floor"),

    ("the priority-pass term limit rises, so three terms are returned",
     [(INBER, "keyTerms = append(keyTerms, textutil.TitleFirstRuneOfEachWord(cleanWord))\n"
       "\t\t\tif len(keyTerms) >= 2 {",
       "keyTerms = append(keyTerms, textutil.TitleFirstRuneOfEachWord(cleanWord))\n"
       "\t\t\tif len(keyTerms) >= 3 {")],
     "key terms: priority limit"),

    ("the backward-scan word floor rises, so three-letter words are skipped",
     [(INBER, "if len(cleanWord) > 2 && !skipWords[cleanWord] && isAlpha(cleanWord) {",
       "if len(cleanWord) > 3 && !skipWords[cleanWord] && isAlpha(cleanWord) {")],
     "key terms: scan floor"),

    ("the backward-scan word floor falls, so two-letter words are accepted",
     [(INBER, "if len(cleanWord) > 2 && !skipWords[cleanWord] && isAlpha(cleanWord) {",
       "if len(cleanWord) > 1 && !skipWords[cleanWord] && isAlpha(cleanWord) {")],
     "key terms: scan floor"),

    ("the backward-scan term limit rises, so three terms are returned",
     [(INBER, "if len(keyTerms) >= 2 { // Limit to 2 key terms for naming",
       "if len(keyTerms) >= 3 { // Limit to 2 key terms for naming")],
     "key terms: scan limit"),

    # -- extractAgentNameFromSpawn: the candidate floor ---------------------
    ("the spawn-name floor rises, so a three-letter agent name is refused",
     [(INBER, "if len(candidate) > 2 && isAlpha(candidate) {",
       "if len(candidate) > 3 && isAlpha(candidate) {")],
     "spawn name floor"),

    ("the spawn-name floor falls, so a two-letter word becomes an agent name",
     [(INBER, "if len(candidate) > 2 && isAlpha(candidate) {",
       "if len(candidate) > 1 && isAlpha(candidate) {")],
     "spawn name floor"),
]

# (label, edits, expect_caught_by_mine, why)
CONTROLS = [
    ("KNOWN-POSITIVE: XP is always zero",
     [(INBER, "return tokens / 100", "return 0")], True,
     "no suite that looks at XP at all can miss every token earning nothing"),

    ("KNOWN-NEGATIVE: the urgent ceiling rises from 5 to 6",
     [(INBER, URGENT_CLAMP, URGENT_CLAMP.replace("min(5,", "min(6,"))], False,
     "dominated: no task type starts above 4 and this clamp runs before every "
     "other bump, so complexity+1 never exceeds 5 here"),

    ("KNOWN-NEGATIVE: the simple-task floor falls from 1 to 0",
     [(INBER, "complexity = max(1, complexity-1)", "complexity = max(0, complexity-1)")], False,
     "dominated: complexity is never below 2 when this runs, so the floor "
     "never lifts anything"),

    ("KNOWN-NEGATIVE: the priority-pass word floor falls from 2 to 1",
     [(INBER, "if priorityWords[cleanWord] && len(cleanWord) > 2 {",
       "if priorityWords[cleanWord] && len(cleanWord) > 1 {")], False,
     "dominated: the shortest word in the priority table is \"api\" at three "
     "letters, so nothing in it can sit between the two floors"),
]


def restore():
    subprocess.run(["git", "checkout", "--"] + TRACKED, cwd=REPO, check=True)


def dirty():
    proc = subprocess.run(["git", "status", "--porcelain"] + TRACKED,
                          cwd=REPO, capture_output=True, text=True)
    return proc.stdout.strip()


def apply(edits):
    """Apply one case's edits. Returns an error string, or None on success."""
    for fname, find, replace in edits:
        path = REPO / fname
        text = path.read_text()
        # Exactly once. Zero matches score UNNOTICED for free; two matches
        # mutate a site the row does not name.
        n = text.count(find)
        if n != 1:
            return "pattern occurs %d times in %s (want exactly 1): %r" % (n, fname, find[:70])
        mutated = text.replace(find, replace, 1)
        if mutated == text:
            return "replacement changed nothing in %s: %r" % (fname, find[:70])
        path.write_text(mutated)
    return None


def run_tests():
    proc = subprocess.run(["go", "test", "-count=1"] + PACKAGES,
                          cwd=REPO, capture_output=True, text=True)
    return verdict_for(proc.stdout + proc.stderr, proc.returncode)


ASIDE = ".prior-baseline"


def move_mine_aside():
    for f in MINE:
        (REPO / f).rename(REPO / (f + ASIDE))


def move_mine_back():
    for f in MINE:
        p = REPO / (f + ASIDE)
        if p.exists():
            p.rename(REPO / f)


def score_case(edits):
    """(prior, mine) verdicts for one mutation."""
    restore()
    err = apply(edits)
    if err:
        return "SETUP FAIL", err

    move_mine_aside()
    try:
        prior = run_tests()
    finally:
        move_mine_back()

    mine = run_tests()
    return prior, mine


if dirty():
    sys.exit("refusing to run: the files under test have uncommitted changes, "
             "and this restores them from git.\n" + dirty())

for f in MINE:
    if not (REPO / f).exists():
        sys.exit("missing test file, so the PRIOR column would be meaningless: " + f)

_previous_handlers = {}


def _restore_and_reraise(signum, frame):
    move_mine_back()
    restore()
    signal.signal(signum, _previous_handlers[signum])
    os.kill(os.getpid(), signum)


for _sig in (signal.SIGINT, signal.SIGTERM, signal.SIGHUP):
    _previous_handlers[_sig] = signal.signal(_sig, _restore_and_reraise)

closed, already, unnoticed, broken = 0, 0, 0, 0
try:
    if self_test() != 0:
        sys.exit("the scorer failed its own self-test")

    print("\n  %-8s %-8s  %-26s %s" % ("PRIOR", "MINE", "site", "mutation"))
    print("  " + "-" * 100)
    for label, edits, site in CASES:
        prior, mine = score_case(edits)
        if prior == "SETUP FAIL":
            print("  SETUP FAIL  %s\n      %s" % (label, mine))
            broken += 1
            continue

        if counts_as_coverage(mine) and not counts_as_coverage(prior):
            mark, closed = "gap closed ", closed + 1
        elif counts_as_coverage(mine) and counts_as_coverage(prior):
            mark, already = "already held", already + 1
        elif not counts_as_coverage(mine):
            mark, unnoticed = "STILL OPEN ", unnoticed + 1
        else:
            mark = "???"
        print("  %-8s %-8s  %-26s %s" % (prior, mine, site, label))
        print("           %s" % mark)

    print("\n  controls")
    print("  " + "-" * 100)
    for label, edits, expect, why in CONTROLS:
        prior, mine = score_case(edits)
        if prior == "SETUP FAIL":
            print("  SETUP FAIL  %s\n      %s" % (label, mine))
            broken += 1
            continue
        ok = counts_as_coverage(mine) == expect
        if not ok:
            broken += 1
        print("  %-8s %-8s  %-26s %s" % (prior, mine, "ok" if ok else "BAD", label))
        print("           %s" % why)
finally:
    move_mine_back()
    restore()
    for _sig, _handler in _previous_handlers.items():
        signal.signal(_sig, _handler)

print("\n  %d gaps closed | %d already held by a neighbour | %d still open | %d broken rows"
      % (closed, already, unnoticed, broken))
sys.exit(1 if broken else 0)
