#!/usr/bin/env python3
"""Sabotage-score the XP reward floor, the last unheld row of a mechanism
written twice.

The sibling scorers each take one mechanism: sabotage-truncation.py and
sabotage-truncation-budgets.py the six cut sites, sabotage-naming-arithmetic.py
the naming and XP ladders, sabotage-journal-tables.py the journal and duration
tables, sabotage-store-paths.py the numbers behind a SQLite fixture,
sabotage-request-boundaries.py the numbers a handler compares a REQUEST against,
sabotage-purefunctions.py the corner of internal/api that needs no fixture, and
sabotage-difficulty-and-pagination.py the http_client half of GetQuests.

This one takes what those left: the XP floor on the STORE path, and the token
sum that feeds it on BOTH paths.

  THE LAST ROW OF A MECHANISM IS THE ONE THE SPLIT LEAVES BEHIND.

GetQuests is written twice, once per data source, and the two bodies are
character-for-character identical through four clusters:

    cluster              inber.go (Store)          http_client.go (HTTPClient)
    difficulty ladder    held (store-paths)        held (difficulty-and-pagination)
    tier values          held (store-paths)        held (difficulty-and-pagination)
    progress constants   held (store-paths)        held (difficulty-and-pagination)
    XP floor             OPEN  <- this scorer      held (difficulty-and-pagination)

Three clusters were closed on both sides by two passes working from opposite
ends. The floor was closed on one side only, because the store-path pass ran
first and nothing in the package asserted XPReward at all until the http pass
existed. So the last row of a four-row mechanism outlived both passes that were
about that mechanism.

  A NEIGHBOUR IS SCORED, NOT READ.

The 208th's rule, and it is what the second half of the case table is for. The
card that filed this work listed xpForTokens' /100 divisor as "already held by
TestXPForTokensCountsTheHundredthTokenAndNotTheNinetyNinth" and asked that it not
be re-scored. Reading a test's name is a claim; a PRIOR/MINE pair is a
measurement, so the divisor is carried here as a neighbour row in both
directions. Same for the three http_client floor rows the sibling closed.

  THE PRIOR COLUMN IS THE WHOLE REPO, NOT THE PACKAGE AND NOT THE FILE.

The 205th's. Every neighbour row below must report PRIOR=CAUGHT -> "already
held". If one reads "gap closed", the base branch is missing a sibling and every
number above it is inflated.

  THE BASE BRANCH IS AN INPUT TO THE SCORE.

The 207th's. Base on test/the-difficulty-ladder-and-pagination-are-unpinned,
which contains all eleven siblings and main. From main, all three http floor
rows read as gaps this pass closed, and the asymmetry the table exists to show
disappears.

  AN IDENTICAL NEEDLE IN TWO FILES MUST BE ANCHORED, NOT COUNTED.

The 211th's, in the same purest form the sibling met it in: `if xpReward < 1 {`
occurs once per file and the two copies differ ONLY by one level of indentation
(three tabs in inber.go, two in http_client.go). Leading tabs are part of every
needle, each row names its file, and apply() counts within that file alone.

  A DOMINATED MUTATION IS A DECLARED CONTROL, NOT AN OPEN GAP.

The 207th's and the 211th's together. `if xpReward < 1 { xpReward = 1 }` spells a
bound one apart from the value it substitutes, so widening it one step to `< 2`
only additionally captures xpReward == 1 and assigns it 1 — the value it already
had. Enumerated over every token count 0..99999 rather than argued:

    N=1 [1,1,1,1,1,1,2,2,3,3]   at tokens 0,50,99,100,150,199,200,250,300,350
    N=2 [1,1,1,1,1,1,2,2,3,3]   IDENTICAL at every one of 100,000 inputs
    N=3 [1,1,1,1,1,1,1,1,3,3]   first differs at 200 tokens, 100 inputs apart

So the upward direction is scored by the TWO-step widening (the 214th's "when ±1
is dominated, try ±2") and `< 2` is carried as a declared known-negative. The
sibling pass filed the same row as a case first and it scored STILL OPEN, which
is indistinguishable from a real gap the next pass should chase.

  A ZERO SUMMAND IS NOT A SUM.

The finding this scorer adds that the card did not predict. Both copies compute
totalTokens from two columns, and EVERY fixture in the package that reaches
either one — oneQuestWithTokens, oneHTTPQuestWithTokens, the ladder and tier
tables, createTestGatewayDB — passes out_tokens 0. A summand that is always zero
is never summed, so the entire output-token column could drop out of the
arithmetic with the whole suite green, on both paths at once. Measured, not
inferred: `grep -rn out_tokens internal/inber/*_test.go` finds no non-zero
value anywhere in the package before this pass.

Rules inherited from earlier passes, each of which cost a session to learn:

  - The needle must occur EXACTLY once in its file and the bytes must change.
  - A compile error is not a red. It is a separate verdict.
  - A panic in a source frame is detection; a panic in a _test.go frame is the
    fixture falling over and is not.
  - Drift the value, do not delete it — including in the CONTROLS, which are
    cases and obey the case rules.
  - Controls in both directions, so a harness that reports CAUGHT for everything
    cannot look perfect.
  - Restores from git, so the tree must be committed before running.

Run:  python3 scripts/sabotage-the-xp-floor.py
      python3 scripts/sabotage-the-xp-floor.py --self-test
      python3 scripts/sabotage-the-xp-floor.py --enumerate-dominance
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
#
# ⚠️ Both halves of this pass's work are in ONE file, including the http_client
# twin that belongs beside its sibling by subject. A test written into a
# committed sibling file would survive the move-aside and report its own row as
# "already held". The PRIOR column is only worth reading if everything this pass
# wrote moves together.
MINE = ["internal/inber/xp_floor_boundaries_test.go"]

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


def quest_xp(tokens, bound):
    """GetQuests' XP arithmetic, with the floor's bound as a parameter.

    Both copies spell it identically: xpForTokens is tokens/100, then the floor
    lifts anything under `bound` to 1.
    """
    xp = tokens // 100
    if xp < bound:
        xp = 1
    return xp


def enumerate_dominance(limit=100000):
    """Prove the one-step widening is dominated instead of arguing it.

    A STILL OPEN row and a dominated mutation look identical in the table, and
    the difference is four lines of enumeration. Returns True if the claims in
    the docstring hold.
    """
    base = [quest_xp(t, 1) for t in range(limit)]
    ok = True
    for bound, want_identical in ((2, True), (3, False), (0, False)):
        mutant = [quest_xp(t, bound) for t in range(limit)]
        differing = [t for t in range(limit) if mutant[t] != base[t]]
        identical = not differing
        if identical != want_identical:
            print("DOMINANCE FAIL: `< %d` identical=%s, want %s" % (bound, identical, want_identical))
            ok = False
        if identical:
            print("  `< %d` vs `< 1`: IDENTICAL at every one of %d inputs — dominated" % (bound, limit))
        else:
            t = differing[0]
            print("  `< %d` vs `< 1`: first differs at %d tokens (%d -> %d), %d differing inputs"
                  % (bound, t, base[t], mutant[t], len(differing)))
    return ok


def self_test():
    """No case table can score its own scorer, so drive every verdict."""
    probes = [
        ("--- FAIL: T\n    x_test.go:9: got 1 want 2\n", 1, "CAUGHT"),
        ("--- FAIL: T\n    x_test.go:9: the fixture no longer reaches GetQuests: served 0 times\n", 1,
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
    # The dominance claim is part of this scorer's correctness: it is why one
    # mutation is a control and not a case. Re-derived on every run.
    if not enumerate_dominance():
        ok = False
    print("self-test: %s" % ("all verdicts reachable and separated" if ok else "FAILED"))
    return 0 if ok else 1


if "--enumerate-dominance" in sys.argv:
    sys.exit(0 if enumerate_dominance() else 1)

if "--self-test" in sys.argv:
    sys.exit(self_test())


# --- Anchors ---------------------------------------------------------------
#
# The floor is byte-identical in the two files apart from one level of
# indentation, so leading tabs are part of every needle and each row also names
# its file. Neither mechanism alone would be enough: the tabs make the needle
# unique repo-wide, the filename makes the ROW unambiguous to a reader.

STORE_FLOOR_GUARD = "\t\t\tif xpReward < 1 {"
STORE_FLOOR_VALUE = "\t\t\t\txpReward = 1"
STORE_TOKEN_SUM = "\t\t\ttotalTokens := inTokens + outTokens"
STORE_XP_FIELD = "\t\t\t\tXPReward:    xpReward,"

HTTP_FLOOR_GUARD = "\t\tif xpReward < 1 {"
HTTP_FLOOR_VALUE = "\t\t\txpReward = 1"
HTTP_TOKEN_SUM = "\t\ttotalTokens := s.InTokens + s.OutTokens"
HTTP_XP_FIELD = "\t\t\tXPReward:    xpReward,"

DIVISOR = "\treturn tokens / 100"


# (label, [(file, find, replace)], site)
#
# Every mutation is a DRIFTED VALUE, not a deletion: a number is far likelier to
# drift by one than a whole check is to disappear, and a drift keeps every
# identifier live so the case reports a score instead of a compile error.
CASES = [
    # -- inber.go: the XP floor, the row this pass exists to close ----------
    # ⚠️ The one-step widening to `< 2` is NOT here. It is dominated — see the
    # known-negative in CONTROLS — so the upward direction is scored by the
    # two-step widening, exactly as the sibling scores the http copy and as the
    # gold guard in internal/api is scored.
    ("the store XP floor guard widens two steps to `< 3`",
     [(STORE, STORE_FLOOR_GUARD, "\t\t\tif xpReward < 3 {")], "store xp floor < 1"),
    ("the store XP floor guard narrows to `< 0`, so a zero reward ships",
     [(STORE, STORE_FLOOR_GUARD, "\t\t\tif xpReward < 0 {")], "store xp floor < 1"),
    ("the store XP floor's substituted value drifts to 2",
     [(STORE, STORE_FLOOR_VALUE, "\t\t\t\txpReward = 2")], "store xp floor value 1"),

    # -- both files: the token sum that feeds the floor ---------------------
    # Not on the card. Every fixture in the package passes out_tokens 0, so the
    # second summand is never summed and can drop out of either copy with the
    # suite green. Scored on both paths because it is one mechanism.
    ("the store token sum drops the output column",
     [(STORE, STORE_TOKEN_SUM, "\t\t\ttotalTokens := inTokens")], "store totalTokens sum"),
    ("the http token sum drops the output column",
     [(HTTP, HTTP_TOKEN_SUM, "\t\ttotalTokens := s.InTokens")], "http totalTokens sum"),

    # -- inber.go: xpForTokens' divisor, carried as a NEIGHBOUR -------------
    # The card said "already held by TestXPForTokensCountsTheHundredth..., listed
    # only so it is not re-scored". Reading a test's name is a claim; this pair
    # of rows is the measurement. Both must read PRIOR=CAUGHT.
    ("NEIGHBOUR: the XP divisor drifts up to /101",
     [(STORE, DIVISOR, "\treturn tokens / 101")], "xpForTokens /100"),
    ("NEIGHBOUR: the XP divisor drifts down to /99",
     [(STORE, DIVISOR, "\treturn tokens / 99")], "xpForTokens /100"),

    # -- http_client.go: the SAME floor, carried as a neighbour -------------
    # Closed by test/the-difficulty-ladder-and-pagination-are-unpinned. All three
    # must read PRIOR=CAUGHT -> "already held". If any reads "gap closed", the
    # base branch is wrong and every number above is inflated.
    ("NEIGHBOUR: the http XP floor guard widens two steps to `< 3`",
     [(HTTP, HTTP_FLOOR_GUARD, "\t\tif xpReward < 3 {")], "http xp floor < 1"),
    ("NEIGHBOUR: the http XP floor guard narrows to `< 0`",
     [(HTTP, HTTP_FLOOR_GUARD, "\t\tif xpReward < 0 {")], "http xp floor < 1"),
    ("NEIGHBOUR: the http XP floor's substituted value drifts to 2",
     [(HTTP, HTTP_FLOOR_VALUE, "\t\t\txpReward = 2")], "http xp floor value 1"),
]


# (label, edits, expected counts_as_coverage, why)
CONTROLS = [
    ("CONTROL: the store XPReward field is carried one too high",
     [(STORE, STORE_XP_FIELD, "\t\t\t\tXPReward:    xpReward + 1,")], True,
     "a positive control on every store row above: the floor, its substituted "
     "value and the token sum are observable ONLY through this one struct "
     "field, so if drifting it does not redden the suite, the store rows are "
     "measuring nothing and their verdicts are theatre. ⚠️ Written as "
     "`xpReward + 1` and not as a constant: replacing the field with a literal "
     "orphans the local and the row scores COMPILE ERROR, which is not a "
     "verdict about coverage. A control is a case and obeys the case rules"),

    ("CONTROL: the http XPReward field is carried one too high",
     [(HTTP, HTTP_XP_FIELD, "\t\t\tXPReward:    xpReward + 1,")], True,
     "the same positive control on the NEIGHBOUR column. The neighbour rows "
     "are only worth reading if the http path is reachable at all under this "
     "base branch; a dead http harness would report every neighbour row "
     "'already held' for the wrong reason and the asymmetry the table exists "
     "to show would be invisible"),

    ("KNOWN-NEGATIVE: the store XP floor guard widens one step to `< 2`",
     [(STORE, STORE_FLOOR_GUARD, "\t\t\tif xpReward < 2 {")], False,
     "DOMINATED, and by the 211th's shape: the guard's bound and the value it "
     "substitutes are one apart, so widening `< 1` to `< 2` only additionally "
     "captures xpReward == 1 and assigns it 1 — the value it already had. "
     "Enumerated over every token count 0..99999 by --enumerate-dominance, "
     "which this scorer's self-test runs on every invocation: `< 2` and `< 1` "
     "agree at EVERY input, and `< 3` first differs at 200 tokens. This row is "
     "carried rather than deleted because it is the row a later pass would "
     "otherwise re-file as a gap: left as a case it scores STILL OPEN, which "
     "is indistinguishable from a real hole"),

    ("KNOWN-NEGATIVE: the http XP floor guard widens one step to `< 2`",
     [(HTTP, HTTP_FLOOR_GUARD, "\t\tif xpReward < 2 {")], False,
     "the same domination in the twin, and it must behave the same way. The "
     "two bodies are character-for-character identical here, so a known-"
     "negative that held in one file and not the other would mean the copies "
     "have diverged — which is exactly the thing this table is built to "
     "notice, and which no test on either side would report"),
]

# ⚠️ NOT SCORED, and declared rather than silently omitted (the 216th's rule):
#
#  1. The difficulty ladder, the tier values and the progress constants. All
#     three are held on BOTH paths already — sabotage-store-paths.py scores the
#     inber.go copies and sabotage-difficulty-and-pagination.py the http ones —
#     and re-scoring a cluster two scorers already own would add nine rows that
#     can only ever read "already held". The floor is here because it was the
#     one cluster of the four that no scorer had on the store side.
#
#  2. The `limit <= 0` substituted default above the floor (inber.go:607). Held
#     by TestGetQuestsSubstitutesFiftyForANonPositiveLimit, and the divergence
#     between the two implementations at limit <= 0 is filed as decision-needed
#     rather than pinned further. Not this pass's call to make.


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
        # hypothetical here, where the floor is spelled twice across two files
        # and differs only in indentation.
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
