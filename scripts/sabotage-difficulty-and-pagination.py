#!/usr/bin/env python3
"""Sabotage-score http_client.go's difficulty ladder, XP floor and pagination.

The sibling scorers each take one mechanism: sabotage-truncation.py and
sabotage-truncation-budgets.py the six cut sites, sabotage-naming-arithmetic.py
the naming and XP ladders, sabotage-journal-tables.py the journal and duration
tables, sabotage-store-paths.py the numbers behind a SQLite fixture,
sabotage-request-boundaries.py the numbers a handler compares a REQUEST against,
and sabotage-purefunctions.py the corner of internal/api that needs no fixture.

This one takes what is left of internal/inber/http_client.go once the
achievements cluster and the 200-byte cut are excluded: the difficulty ladder,
the tier values it assigns, the XP floor, the progress constants, and the two
pagination guards.

  THE MECHANISM SPANS TWO FILES, AND THE SPLIT CUT IT DOWN THE MIDDLE.

The finding this scorer exists to make legible. GetQuests is written TWICE:

    internal/inber/inber.go:685        (*Store).GetQuests       SQLite-backed
    internal/inber/http_client.go:180  (*HTTPClient).GetQuests  HTTP-backed

The difficulty ladder, the XP floor and the progress constants are
character-for-character identical in the two bodies. sabotage-store-paths.py and
store_path_boundaries_test.go closed the inber.go copy. Nothing reached the
http_client.go copy, and a grep for "difficulty ladder" finds the sibling test
and reports the mechanism covered.

The onboarding note `e72c6e27` argued this exact point about the TRUNCATION
family and said "split by mechanism, not by file". The difficulty family is a
second mechanism with the same shape and nobody noticed, because the two copies
sit behind two different data sources rather than in five obviously-different
functions.

So the inber.go rows below are carried DELIBERATELY, as neighbours:

  THE PRIOR COLUMN IS THE WHOLE REPO, NOT THE PACKAGE AND NOT THE FILE.

Inherited from the 205th pass. Every inber.go row must report PRIOR=CAUGHT ->
"already held". If one ever reads "gap closed", the base branch is missing
test/the-store-path-tiers-are-unpinned and every http_client.go number here is
inflated. Scoring the twin is also the only way the asymmetry is visible in one
table: identical mutation, two files, two different PRIOR columns.

  THE BASE BRANCH IS AN INPUT TO THE SCORE.

The 207th's. Base on test/the-pure-function-boundaries-are-unpinned, which
contains all ten siblings and main. From main every neighbour row reads as a gap
this pass closed, and the whole point of carrying them is lost.

  AN IDENTICAL NEEDLE IN TWO FILES MUST BE ANCHORED, NOT COUNTED.

The 211th's rule, met here in its purest form: `if totalTokens > 5000 {` occurs
once per file, so a repo-wide needle would be ambiguous and a per-file needle is
not. Every row below names its file, and apply() counts within that file only.
The two copies differ solely in indentation (two tabs in http_client.go, three
in inber.go), which is not a difference any reader should be asked to rely on.

  A DOMINATED MUTATION IS A DECLARED CONTROL, NOT AN OPEN GAP.

The 207th's, and two rows here are dominated for two different reasons:

  1. BY THE SHAPE OF THE LOOP, not by a constant. GetQuests appends
     unconditionally, so at every iteration `len(quests) == i` and rewriting
     `i >= limit` as `len(quests) >= limit` is the identity. Carried as a
     known-negative so that a future `continue` added to that loop turns it into
     a real gap loudly. Its neighbour GetQuestHistory DOES have a `continue`,
     so the same rewrite there is a real defect — the same edit is the identity
     in one function and a bug in the other.

  2. BY THE GUARD BEING ONE APART FROM THE VALUE IT SUBSTITUTES. This is the
     211th's shape, met again in a different file: `if xpReward < 1 { xpReward
     = 1 }` cannot be widened one step observably, because `< 2` only adds the
     case xpReward == 1, which is assigned 1 anyway. `< 3` IS observable, at
     200 tokens. Enumerated over 0..99999 rather than argued.

     ⚠️ It was a case first and scored STILL OPEN, which reads exactly like a
     gap the next pass should close. The 214th's "when ±1 is dominated, try ±2"
     is what turned it into a scored pair plus a control instead.

  AN UNREACHABLE BRANCH IS STILL A BRANCH, AND ITS TEST SAYS SO.

Both `limit > 0` guards are observable at exactly one input, limit == 0, and no
production caller passes 0: api.go's two handlers substitute their defaults
unless the parsed limit is `> 0`, and the four in-package callers pass 1000 or
100 literally. The rows are scored anyway — the function is exported and its
contract is what a test pins — but the tests say in a comment that the branch
does not ship, so the next reader does not infer coverage of a live path.

Rules inherited from earlier passes, each of which cost a session to learn:

  - The needle must occur EXACTLY once in its file and the bytes must change.
  - A compile error is not a red. It is a separate verdict.
  - A panic in a source frame is detection; a panic in a _test.go frame is the
    fixture falling over and is not.
  - Controls in both directions, so a harness that reports CAUGHT for everything
    cannot look perfect.
  - Restores from git, so the tree must be committed before running.

Run:  python3 scripts/sabotage-difficulty-and-pagination.py
      python3 scripts/sabotage-difficulty-and-pagination.py --self-test
"""

import os
import pathlib
import re
import signal
import subprocess
import sys

REPO = pathlib.Path(__file__).resolve().parent.parent

HTTP = "internal/inber/http_client.go"
STORE = "internal/inber/inber.go"

TRACKED = [HTTP, STORE]

# The whole repo, for the reason in the docstring.
PACKAGES = ["./..."]

# The file this sweep adds. Renamed aside to build the PRIOR column.
MINE = ["internal/inber/http_client_boundaries_test.go"]

# Reach guards, by message. A guard firing means a fixture stopped reaching the
# code, which is not the same as an assertion noticing a changed value.
GUARD_MARKERS = (
    'input no longer reaches the truncating branch',
    "no longer reaches generateQuestName's truncating fallback",
    'is missing from this working tree',
    'cannot tell which arm produced',
    'the fixture no longer reaches GetQuests',
    'the fixture no longer reaches GetQuestHistory',
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
        ("--- FAIL: T\n    x_test.go:9: the fixture no longer reaches GetQuests: served 0 times\n", 1,
         "CAUGHT (guard)"),
        ("--- FAIL: T\n    x_test.go:9: the fixture no longer reaches GetQuestHistory: served 0\n", 1,
         "CAUGHT (guard)"),
        ("--- FAIL: T\n    x_test.go:9: the fixture no longer reaches GetQuests: served 0 times\n"
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


# --- Anchors ---------------------------------------------------------------
#
# The ladder, the floor and the progress block are byte-identical in the two
# files apart from one level of indentation. Leading tabs are therefore part of
# every needle, and each row also names its file so apply() counts in one file
# only. Neither mechanism alone would be enough: the tabs make the needle unique
# repo-wide, the filename makes the ROW unambiguous to a reader.

HTTP_LADDER = {
    5000: "\t\tif totalTokens > 5000 {",
    2000: "\t\t} else if totalTokens > 2000 {",
    1000: "\t\t} else if totalTokens > 1000 {",
    500: "\t\t} else if totalTokens > 500 {",
}

STORE_LADDER = {
    5000: "\t\t\tif totalTokens > 5000 {",
    2000: "\t\t\t} else if totalTokens > 2000 {",
    1000: "\t\t\t} else if totalTokens > 1000 {",
    500: "\t\t\t} else if totalTokens > 500 {",
}

HTTP_TIER = {
    5: "\t\t\tdifficulty = 5",
    4: "\t\t\tdifficulty = 4",
    3: "\t\t\tdifficulty = 3",
    2: "\t\t\tdifficulty = 2",
    1: "\t\tdifficulty := 1",
}

HTTP_XP_FLOOR = "\t\tif xpReward < 1 {"
HTTP_XP_FLOOR_VALUE = "\t\t\txpReward = 1"

HTTP_PROGRESS = {
    50: "\t\tprogress := 50",
    100: "\t\t\tprogress = 100",
    30: "\t\t\tprogress = 30",
    0: "\t\t\tprogress = 0",
}

QUESTS_CUT = "\t\tif limit > 0 && i >= limit {"
HISTORY_CUT = "\t\tif limit > 0 && len(entries) >= limit {"


def threshold(table, fname, value, replacement):
    """One ladder row, anchored on its own indentation and file."""
    find = table[value]
    return [(fname, find, find.replace("> %d" % value, "> %s" % replacement, 1))]


# (label, [(file, find, replace)], site)
#
# Every mutation is a DRIFTED VALUE, not a deletion: a number is far likelier to
# drift by one than a whole check is to disappear, and a drift keeps every
# identifier live so the case reports a score instead of a compile error.
CASES = [
    # -- http_client.go: the difficulty ladder's four thresholds ------------
    # Each is scored in BOTH directions. A single-sided table would leave the
    # number free to move the other way with every test still green, which is
    # the specific failure the straddle pairs in the test file exist to close.
    ("the legendary threshold drifts up to 5001",
     threshold(HTTP_LADDER, HTTP, 5000, 5001), "http ladder > 5000"),
    ("the legendary threshold drifts down to 4999",
     threshold(HTTP_LADDER, HTTP, 5000, 4999), "http ladder > 5000"),

    ("the expert threshold drifts up to 2001",
     threshold(HTTP_LADDER, HTTP, 2000, 2001), "http ladder > 2000"),
    ("the expert threshold drifts down to 1999",
     threshold(HTTP_LADDER, HTTP, 2000, 1999), "http ladder > 2000"),

    ("the hard threshold drifts up to 1001",
     threshold(HTTP_LADDER, HTTP, 1000, 1001), "http ladder > 1000"),
    ("the hard threshold drifts down to 999",
     threshold(HTTP_LADDER, HTTP, 1000, 999), "http ladder > 1000"),

    ("the medium threshold drifts up to 501",
     threshold(HTTP_LADDER, HTTP, 500, 501), "http ladder > 500"),
    ("the medium threshold drifts down to 499",
     threshold(HTTP_LADDER, HTTP, 500, 499), "http ladder > 500"),

    # -- http_client.go: the tier VALUES the ladder assigns -----------------
    # A separate question from the thresholds: 5/4/3/2/1 and 6/4/3/2/1 cross at
    # exactly the same token counts, so the straddle table cannot see these.
    ("the legendary tier value drifts to 6",
     [(HTTP, HTTP_TIER[5], "\t\t\tdifficulty = 6")], "http tier value 5"),
    ("the expert tier value drifts to 6",
     [(HTTP, HTTP_TIER[4], "\t\t\tdifficulty = 6")], "http tier value 4"),
    ("the hard tier value drifts to 6",
     [(HTTP, HTTP_TIER[3], "\t\t\tdifficulty = 6")], "http tier value 3"),
    ("the medium tier value drifts to 6",
     [(HTTP, HTTP_TIER[2], "\t\t\tdifficulty = 6")], "http tier value 2"),
    ("the trivial tier default drifts to 0",
     [(HTTP, HTTP_TIER[1], "\t\tdifficulty := 0")], "http tier default 1"),

    # -- http_client.go: the XP floor --------------------------------------
    # Unheld in BOTH files before this sweep: nothing in the package asserted
    # XPReward at all.
    # ⚠️ The one-step widening to `< 2` is NOT here. It is dominated — see the
    # known-negative in CONTROLS — so the upward direction is scored by the
    # two-step widening below, exactly as the gold guard in internal/api is.
    ("the XP floor guard widens two steps to `< 3`",
     [(HTTP, HTTP_XP_FLOOR, "\t\tif xpReward < 3 {")], "http xp floor < 1"),
    ("the XP floor guard narrows to `< 0`, so a zero reward ships",
     [(HTTP, HTTP_XP_FLOOR, "\t\tif xpReward < 0 {")], "http xp floor < 1"),
    ("the XP floor's substituted value drifts to 2",
     [(HTTP, HTTP_XP_FLOOR_VALUE, "\t\t\txpReward = 2")], "http xp floor value 1"),

    # -- http_client.go: the progress constants ----------------------------
    ("the in-progress constant drifts to 51",
     [(HTTP, HTTP_PROGRESS[50], "\t\tprogress := 51")], "http progress 50"),
    ("the completed constant drifts to 99",
     [(HTTP, HTTP_PROGRESS[100], "\t\t\tprogress = 99")], "http progress 100"),
    ("the failed constant drifts to 31",
     [(HTTP, HTTP_PROGRESS[30], "\t\t\tprogress = 31")], "http progress 30"),
    ("the available constant drifts to 1",
     [(HTTP, HTTP_PROGRESS[0], "\t\t\tprogress = 1")], "http progress 0"),

    # -- http_client.go: GetQuests pagination ------------------------------
    ("the quest cut lands one late (`i > limit`)",
     [(HTTP, QUESTS_CUT, "\t\tif limit > 0 && i > limit {")], "http GetQuests cut"),
    ("the quest cut lands one early (`i >= limit-1`)",
     [(HTTP, QUESTS_CUT, "\t\tif limit > 0 && i >= limit-1 {")], "http GetQuests cut"),
    ("the quest cut's `limit > 0` widens to `>= 0`, so limit=0 returns nothing",
     [(HTTP, QUESTS_CUT, "\t\tif limit >= 0 && i >= limit {")], "http GetQuests limit > 0"),

    # -- http_client.go: GetQuestHistory pagination ------------------------
    ("the history cut lands one late (`> limit`)",
     [(HTTP, HISTORY_CUT, "\t\tif limit > 0 && len(entries) > limit {")],
     "http GetQuestHistory cut"),
    ("the history cut lands one early (`>= limit-1`)",
     [(HTTP, HISTORY_CUT, "\t\tif limit > 0 && len(entries) >= limit-1 {")],
     "http GetQuestHistory cut"),
    ("the history cut's `limit > 0` widens to `>= 0`, so limit=0 returns one",
     [(HTTP, HISTORY_CUT, "\t\tif limit >= 0 && len(entries) >= limit {")],
     "http GetQuestHistory limit > 0"),

    # -- inber.go: the SAME ladder, carried as a neighbour -----------------
    # These must all read PRIOR=CAUGHT -> "already held". They are the other
    # half of the mechanism, closed by test/the-store-path-tiers-are-unpinned.
    # If any reads "gap closed", the base branch is wrong and every number above
    # is inflated. Both directions, so the neighbour is checked as strictly as
    # the new work.
    ("NEIGHBOUR: the store ladder's legendary threshold drifts up to 5001",
     threshold(STORE_LADDER, STORE, 5000, 5001), "store ladder > 5000"),
    ("NEIGHBOUR: the store ladder's legendary threshold drifts down to 4999",
     threshold(STORE_LADDER, STORE, 5000, 4999), "store ladder > 5000"),
    ("NEIGHBOUR: the store ladder's medium threshold drifts up to 501",
     threshold(STORE_LADDER, STORE, 500, 501), "store ladder > 500"),
]


# (label, edits, expected counts_as_coverage, why)
CONTROLS = [
    ("CONTROL: the quest cut is deleted outright",
     [(HTTP, QUESTS_CUT, "\t\tif false && i >= limit {")], True,
     "a positive control on the pagination rows. If deleting the cut entirely "
     "does not redden the suite, the limit tables are asserting nothing and "
     "every pagination row above is theatre"),

    ("CONTROL: the difficulty field is carried one tier too high",
     [(HTTP, "\t\t\tDifficulty:  difficulty,", "\t\t\tDifficulty:  difficulty + 1,")], True,
     "a positive control on the whole ladder cluster: the thresholds and tier "
     "values are only observable through this one struct field, so if pinning "
     "it does not redden, the ladder rows are measuring the wrong thing. "
     "⚠️ Written as `difficulty + 1` and not as a constant: replacing the "
     "field with `1` orphans the local and the row scores COMPILE ERROR, which "
     "is not a verdict about coverage. This is the case-table rule ('drift the "
     "value, do not delete it') applying to controls as well, and it cost this "
     "pass one broken row to notice"),

    ("KNOWN-NEGATIVE: the XP floor guard widens one step to `< 2`",
     [(HTTP, HTTP_XP_FLOOR, "\t\tif xpReward < 2 {")], False,
     "DOMINATED, and by the 211th's shape: the guard's bound and the value it "
     "substitutes are one apart, so widening `< 1` to `< 2` only additionally "
     "captures xpReward == 1 and assigns it 1 — the value it already had. "
     "Enumerated over every token count 0..99999: `< 2` and `< 1` agree at "
     "EVERY input, and `< 3` first differs at 200 tokens. So the upward "
     "direction is pinned by the two-step case above and not by this one. "
     "⚠️ This row was a case first and scored STILL OPEN, which is "
     "indistinguishable from a real gap; the 211th's rule is to enumerate the "
     "mutant against the original before leaving such a row standing, and "
     "doing so is what turned it into a control"),

    ("KNOWN-NEGATIVE: the quest cut counts appends instead of indices",
     [(HTTP, QUESTS_CUT, "\t\tif limit > 0 && len(quests) >= limit {")], False,
     "DOMINATED by the loop's own shape, not by a constant. GetQuests appends "
     "unconditionally — there is no `continue` between the guard and the "
     "append — so len(quests) == i at every iteration and the rewrite is the "
     "identity. Enumerated over limits 0..7 against a five-session fixture "
     "before it was declared. ⚠️ This row is carried rather than deleted so "
     "that adding a `continue` to that loop turns it into a real gap loudly, "
     "instead of silently making two pinned behaviours disagree"),

]

# ⚠️ NOT SCORED, and declared rather than silently omitted (the 216th's rule):
# GetQuestHistory has no matching known-negative. Its loop DOES carry a
# `continue` — the agent filter — so there len(entries) and the iteration index
# genuinely diverge and swapping them is a real behaviour change, already
# covered by the case rows above and by
# TestGetQuestHistoryReturnsOnlyTheNamedAgentsQuests. The asymmetry between the
# two loops is the reason the GetQuests row needed declaring at all: the same
# rewrite is the identity in one function and a defect in its neighbour.


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
        # Exactly once IN THIS FILE. Zero matches score UNNOTICED for free; two
        # matches mutate a site the row does not name — which is not
        # hypothetical here, where the whole ladder is spelled twice across two
        # files and differs only in indentation.
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

    print("\n  %-8s %-8s  %-34s %s" % ("PRIOR", "MINE", "site", "mutation"))
    print("  " + "-" * 108)
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
        print("  %-8s %-8s  %-34s %s" % (prior, mine, site, label))
        print("           %s" % mark)

    print("\n  controls")
    print("  " + "-" * 108)
    for label, edits, expect, why in CONTROLS:
        prior, mine = score_case(edits)
        if prior == "SETUP FAIL":
            print("  SETUP FAIL  %s\n      %s" % (label, mine))
            broken += 1
            continue
        ok = counts_as_coverage(mine) == expect
        if not ok:
            broken += 1
        print("  %-8s %-8s  %-34s %s" % (prior, mine, "ok" if ok else "BAD", label))
        print("           %s" % why)
finally:
    move_mine_back()
    restore()
    for _sig, _handler in _previous_handlers.items():
        signal.signal(_sig, _handler)

print("\n  %d gaps closed | %d already held by a neighbour | %d still open | %d broken rows"
      % (closed, already, unnoticed, broken))
sys.exit(1 if broken else 0)
