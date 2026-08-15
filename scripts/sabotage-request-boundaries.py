#!/usr/bin/env python3
"""Sabotage-score api.go's REQUEST-VALIDATED boundaries.

The sibling scorers cover what happens to data once a handler has decided to
act: sabotage-truncation.py and sabotage-truncation-budgets.py score the six cut
sites, sabotage-naming-arithmetic.py the naming and XP arithmetic,
sabotage-journal-tables.py the journal and duration tables, and
sabotage-store-paths.py the numbers behind a SQLite fixture.

This one scores the numbers a handler compares a REQUEST against, before any of
that: the three pagination blocks' defaults and caps, the two ValidateLimit call
sites' arguments, the YYYY-MM-DD length check, and the required-skills ceiling.
Every case drives a real handler through httptest, and every one of them was
free before request_validated_boundaries_test.go existed, because api_test.go
never sent a query string at all.

  THE PRIOR COLUMN IS THE WHOLE REPO, NOT THE PACKAGE AND NOT THE FILE.

Inherited from the 205th pass. This sweep has a live sibling in the same
package: spawn_message_budget_test.go holds api.go's spawn-message cut, which
is in this file too. Two rows below are carried for exactly that reason and are
EXPECTED to report PRIOR=CAUGHT -> "already held". If either ever reads
"gap closed", the base branch is wrong and every other number here is inflated.
request_validated_boundaries_test.go asserts the same premise directly, so that
failure mode shows up as a plain test failure as well.

  THE BASE BRANCH IS AN INPUT TO THE SCORE.

The 207th's. Run from `main` and the two neighbour rows read as gaps this pass
closed. Base on test/the-store-path-tiers-are-unpinned, which merges all eight
siblings, or on a branch that merges test/the-truncation-budgets-are-unpinned.

  AN UNREACHABLE MUTATION IS A DECLARED CONTROL, NOT AN OPEN GAP.

The 207th's dominance rule, and this sweep's largest single finding.
api.go:2536-2540 computes a default bounty tier from a 1000/500/100 ladder.
db.CreateBounty's last act (bounties.go:143) is `b.Tier = tier`, recomputed
unconditionally from its OWN 500/200/50 ladder, over the same pointer, before
api.go encodes the response. So api.go's ladder is dead: measured, replacing the
whole thing with `bounty.Tier = "COMPLETELY-BOGUS-TIER"` leaves every response
and every stored row byte-identical. Its three thresholds and four tier strings
are carried below as declared KNOWN-NEGATIVES. Filing them as gaps would send
the next pass after a test that cannot exist.

  A NEEDLE MUST NAME ONE SITE, AND AN ANCHORLESS ONE HERE WILL NOT.

`if limit > 100 { // Cap at 100` and its `limit = 100` occur TWICE in api.go —
handleConversations (:1513) and listDisputes (:2930) are the same block
copy-pasted, differing only in the default above them. A bare needle mutates the
first and reports it under whichever row asked. Both sites are anchored on the
`limit := N` line that distinguishes them. This is also why the sweep covers
listDisputes at all: the grep that sized the card saw one line where there were
two sites.

Rules inherited from earlier passes, each of which cost a session to learn:

  - The needle must occur EXACTLY once and the file's bytes must change.
  - A compile error is not a red. It is a separate verdict.
  - A panic in a source frame is detection; a panic in a _test.go frame is the
    fixture falling over and is not.
  - Controls in both directions, so a harness that reports CAUGHT for
    everything cannot look perfect.
  - Restores from git, so the tree must be committed before running.

Run:  python3 scripts/sabotage-request-boundaries.py
      python3 scripts/sabotage-request-boundaries.py --self-test
"""

import os
import pathlib
import re
import signal
import subprocess
import sys

REPO = pathlib.Path(__file__).resolve().parent.parent

API = "internal/api/api.go"
LOGSTACK = "internal/logstack/logstack.go"

TRACKED = [API, LOGSTACK]

# The whole repo, for the reason in the docstring.
PACKAGES = ["./..."]

# The file this sweep adds. Renamed aside to build the PRIOR column.
MINE = ["internal/api/request_validated_boundaries_test.go"]

# Reach guards, by message. Inherited from the truncation tests, which share
# this package. This sweep contributes one: the base-branch premise check, which
# is a guard rather than a measurement of any boundary.
GUARD_MARKERS = (
    'input no longer reaches the truncating branch',
    "no longer reaches generateQuestName's truncating fallback",
    'is missing from this working tree',
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
        ("--- FAIL: T\n    x_test.go:9: spawn_message_budget_test.go is missing from this "
         "working tree: no such file\n", 1, "CAUGHT (guard)"),
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


# --- Anchors for the pagination block api.go spells three times -------------
#
# Written out here rather than inline so a reader can check that the anchor
# names the site the row claims. handleConversations and listDisputes are
# character-for-character identical from `if limitStr != ""` down, so only the
# `limit := N` line above separates them, and listBounties reaches the same
# ceiling through ValidateLimit instead.

CONVERSATIONS_BLOCK = '''\tlimit := 20
\tif limitStr != "" {
\t\tif parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
\t\t\tlimit = parsed
\t\t\tif limit > 100 { // Cap at 100
\t\t\t\tlimit = 100
\t\t\t}
\t\t}
\t}'''

DISPUTES_BLOCK = '''\tlimit := 50
\tif limitStr != "" {
\t\tif parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
\t\t\tlimit = parsed
\t\t\tif limit > 100 { // Cap at 100
\t\t\t\tlimit = 100
\t\t\t}
\t\t}
\t}'''

# logstack's own substituted default, spelled 20 like the handler's. It is
# unreachable THROUGH this handler, which always passes a positive limit, and it
# belongs to the logstack child of this sweep rather than to this file. Carried
# as a known-negative so that claim is measured rather than asserted in prose.
LOGSTACK_DEFAULT = ('func (lc *LogstackClient) GetAgentConversations(agentID string, '
                    'limit int) ([]Conversation, error) {\n'
                    '\tif limit <= 0 {\n'
                    '\t\tlimit = 20')


def block_case(block, find, replace):
    """One edit inside an anchored block, so the row names exactly one site."""
    assert block.count(find) == 1, "ambiguous inside the block: %r" % (find,)
    return [(API, block, block.replace(find, replace, 1))]


# (label, [(file, find, replace)], site)
#
# Every mutation is a DRIFTED VALUE, not a deletion: a number is far likelier to
# drift by one than a whole check is to disappear, and a drift keeps every
# identifier live so the case reports a score instead of a compile error.
CASES = [
    # -- handleConversations: default, cap guard, cap value -----------------
    # The default and the ceiling are separate numbers, and the ceiling is
    # spelled twice (the guard that decides to cut, the value it cuts to), so
    # the block yields three independent boundaries and six rows.
    ("the conversation default limit drifts down to 19",
     block_case(CONVERSATIONS_BLOCK, "limit := 20", "limit := 19"),
     "conversations default 20"),

    ("the conversation default limit drifts up to 21",
     block_case(CONVERSATIONS_BLOCK, "limit := 20", "limit := 21"),
     "conversations default 20"),

    ("the conversation cap GUARD drifts up, so 101 is no longer cut",
     block_case(CONVERSATIONS_BLOCK, "if limit > 100 {", "if limit > 101 {"),
     "conversations cap guard"),

    # The GUARD's other direction is DOMINATED and is scored as a control
    # below, not as a case. See the note there.

    ("the conversation cap VALUE drifts up to 101",
     block_case(CONVERSATIONS_BLOCK, "\t\t\t\tlimit = 100", "\t\t\t\tlimit = 101"),
     "conversations cap value"),

    ("the conversation cap VALUE drifts down to 99",
     block_case(CONVERSATIONS_BLOCK, "\t\t\t\tlimit = 100", "\t\t\t\tlimit = 99"),
     "conversations cap value"),

    # -- listDisputes: the same block, at the site the card did not name ----
    ("the dispute default limit drifts down to 49",
     block_case(DISPUTES_BLOCK, "limit := 50", "limit := 49"), "disputes default 50"),

    ("the dispute default limit drifts up to 51",
     block_case(DISPUTES_BLOCK, "limit := 50", "limit := 51"), "disputes default 50"),

    ("the dispute cap GUARD drifts up, so 101 is no longer cut",
     block_case(DISPUTES_BLOCK, "if limit > 100 {", "if limit > 101 {"),
     "disputes cap guard"),

    ("the dispute cap VALUE drifts down to 99",
     block_case(DISPUTES_BLOCK, "\t\t\t\tlimit = 100", "\t\t\t\tlimit = 99"),
     "disputes cap value"),

    # -- listCosts: ValidateLimit's ARGUMENTS ------------------------------
    # The 204th's point about TruncateAtRuneBoundary(task, 100): a literal
    # passed as a call argument is invisible to the grep that sized this sweep.
    # Both numbers below are call arguments. validation_test.go:460 pins how
    # ValidateLimit behaves given (10, 100) and says nothing about what any
    # caller passes, which is the number that ships.
    ("the cost default limit drifts down to 99",
     [(API, 'ValidateLimit(r.URL.Query().Get("limit"), 100, 1000)',
       'ValidateLimit(r.URL.Query().Get("limit"), 99, 1000)')],
     "costs default limit 100"),

    ("the cost default limit drifts up to 101",
     [(API, 'ValidateLimit(r.URL.Query().Get("limit"), 100, 1000)',
       'ValidateLimit(r.URL.Query().Get("limit"), 101, 1000)')],
     "costs default limit 100"),

    ("the cost max limit drifts down to 999",
     [(API, 'ValidateLimit(r.URL.Query().Get("limit"), 100, 1000)',
       'ValidateLimit(r.URL.Query().Get("limit"), 100, 999)')],
     "costs max limit 1000"),

    ("the cost max limit drifts up to 1001",
     [(API, 'ValidateLimit(r.URL.Query().Get("limit"), 100, 1000)',
       'ValidateLimit(r.URL.Query().Get("limit"), 100, 1001)')],
     "costs max limit 1000"),

    # -- listBounties: the OTHER ValidateLimit call site --------------------
    # Different arguments, same function, same file. A test pinning one call
    # site leaves the other free, which is why both are scored.
    ("the bounty default limit drifts down to 49",
     [(API, 'ValidateLimit(r.URL.Query().Get("limit"), 50, 100)',
       'ValidateLimit(r.URL.Query().Get("limit"), 49, 100)')],
     "bounties default limit 50"),

    ("the bounty default limit drifts up to 51",
     [(API, 'ValidateLimit(r.URL.Query().Get("limit"), 50, 100)',
       'ValidateLimit(r.URL.Query().Get("limit"), 51, 100)')],
     "bounties default limit 50"),

    ("the bounty max limit drifts down to 99",
     [(API, 'ValidateLimit(r.URL.Query().Get("limit"), 50, 100)',
       'ValidateLimit(r.URL.Query().Get("limit"), 50, 99)')],
     "bounties max limit 100"),

    ("the bounty max limit drifts up to 101",
     [(API, 'ValidateLimit(r.URL.Query().Get("limit"), 50, 100)',
       'ValidateLimit(r.URL.Query().Get("limit"), 50, 101)')],
     "bounties max limit 100"),

    # -- listCosts: the YYYY-MM-DD LENGTH check -----------------------------
    # `len(date) != 10` is a string-length boundary, not a parse. Only a
    # 9/10/11 straddle separates it, and each parameter is checked separately
    # so a drift on one alone stays visible.
    ("the start_date length requirement drifts to 9",
     [(API, "len(startDate) != 10", "len(startDate) != 9")], "start_date length 10"),

    ("the start_date length requirement drifts to 11",
     [(API, "len(startDate) != 10", "len(startDate) != 11")], "start_date length 10"),

    ("the end_date length requirement drifts to 9",
     [(API, "len(endDate) != 10", "len(endDate) != 9")], "end_date length 10"),

    ("the end_date length requirement drifts to 11",
     [(API, "len(endDate) != 10", "len(endDate) != 11")], "end_date length 10"),

    # -- createBounty: the required-skills ceiling --------------------------
    ("the required-skills ceiling tightens, so 10 skills are rejected",
     [(API, "len(bounty.RequiredSkills) > 10", "len(bounty.RequiredSkills) > 9")],
     "required skills ceiling 10"),

    ("the required-skills ceiling loosens, so 11 skills are accepted",
     [(API, "len(bounty.RequiredSkills) > 10", "len(bounty.RequiredSkills) > 11")],
     "required skills ceiling 10"),

    # -- The spawn-message cut, scored to PROVE it is a neighbour's ---------
    # The 208th's rule: when a card's scope names something a neighbour already
    # covers, put it in the case table and let it score `already held`. Do not
    # delete it with a comment — if the reading is wrong, or the base branch is
    # missing that neighbour, the number just comes out smaller and nothing
    # catches it. These two rows are expected to report PRIOR=CAUGHT. If either
    # reads "gap closed", the base is wrong and every number above is inflated.
    ("the spawn message cut's GUARD drifts up by one",
     [(API, "if len(task) > 100 {", "if len(task) > 101 {")],
     "spawn cut guard (neighbour)"),

    ("the spawn message cut's BUDGET drifts down by one",
     [(API, "TruncateAtRuneBoundary(task, 100)", "TruncateAtRuneBoundary(task, 99)")],
     "spawn cut budget (neighbour)"),
]

# (label, edits, expect_caught_by_mine, why)
CONTROLS = [
    ("KNOWN-POSITIVE: the conversation cap collapses to 5",
     block_case(CONVERSATIONS_BLOCK, "\t\t\t\tlimit = 100", "\t\t\t\tlimit = 5"), True,
     "this sweep asserts a 105-file agent returns exactly 100 at three limits, "
     "so a harness that cannot report CAUGHT here is dead"),

    ("KNOWN-POSITIVE: the required-skills ceiling is removed",
     [(API, "len(bounty.RequiredSkills) > 10", "len(bounty.RequiredSkills) > 1000")], True,
     "11 skills would be accepted; this sweep asserts they are rejected"),

    ("KNOWN-NEGATIVE: api.go's legendary tier threshold drifts to 1001",
     [(API, "if bounty.PayoutAmount >= 1000 {", "if bounty.PayoutAmount >= 1001 {")], False,
     "DEAD CODE, not a fixture gap: db.CreateBounty recomputes b.Tier from its "
     "own 500/200/50 ladder over the same pointer before api.go encodes the "
     "response, so nothing ever reads what this branch assigns"),

    ("KNOWN-NEGATIVE: api.go's silver tier threshold drifts to 101",
     [(API, "} else if bounty.PayoutAmount >= 100 {", "} else if bounty.PayoutAmount >= 101 {")],
     False,
     "same reason: overwritten before it is read. Scored so that the dead-code "
     "claim is a measurement and not a sentence in a comment"),

    ("KNOWN-NEGATIVE: api.go's legendary tier STRING drifts to mythic",
     [(API, '\t\t\tbounty.Tier = "legendary"', '\t\t\tbounty.Tier = "mythic"')], False,
     "the strings are as dead as the thresholds; if this ever reports CAUGHT, "
     "CreateBounty stopped overwriting the tier and api.go's ladder is live "
     "again — at which point four more rows become real gaps"),

    ("KNOWN-NEGATIVE: the conversation cap guard drifts DOWN to `> 99`",
     block_case(CONVERSATIONS_BLOCK, "if limit > 100 {", "if limit > 99 {"), False,
     "DOMINATED, not an open gap: the guard and the value it cuts to are the "
     "SAME number, so loosening the guard by one only pulls in limit=100, which "
     "the branch then assigns 100 — the value it already had. Enumerated over "
     "99/100/101/500 the mutated handler returns an identical answer at every "
     "input, so no test can separate it. The 207th's rule: score the "
     "unreachable direction as a declared control. This is why the sweep pins "
     "each ceiling from ONE side and pins the VALUE separately, and it applies "
     "verbatim to listDisputes' identical block"),

    ("KNOWN-NEGATIVE: logstack's own substituted default drifts to 19",
     [(LOGSTACK, LOGSTACK_DEFAULT, LOGSTACK_DEFAULT.replace("limit = 20", "limit = 19"))],
     False,
     "unreachable through handleConversations, which only ever calls through "
     "with a positive limit. It belongs to the logstack child of this sweep, "
     "and is scored here to prove this file does not silently claim it"),
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
        # mutate a site the row does not name — which is not hypothetical in
        # this file, where the pagination block occurs three times.
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

    print("\n  %-8s %-8s  %-30s %s" % ("PRIOR", "MINE", "site", "mutation"))
    print("  " + "-" * 104)
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
        print("  %-8s %-8s  %-30s %s" % (prior, mine, site, label))
        print("           %s" % mark)

    print("\n  controls")
    print("  " + "-" * 104)
    for label, edits, expect, why in CONTROLS:
        prior, mine = score_case(edits)
        if prior == "SETUP FAIL":
            print("  SETUP FAIL  %s\n      %s" % (label, mine))
            broken += 1
            continue
        ok = counts_as_coverage(mine) == expect
        if not ok:
            broken += 1
        print("  %-8s %-8s  %-30s %s" % (prior, mine, "ok" if ok else "BAD", label))
        print("           %s" % why)
finally:
    move_mine_back()
    restore()
    for _sig, _handler in _previous_handlers.items():
        signal.signal(_sig, _handler)

print("\n  %d gaps closed | %d already held by a neighbour | %d still open | %d broken rows"
      % (closed, already, unnoticed, broken))
sys.exit(1 if broken else 0)
