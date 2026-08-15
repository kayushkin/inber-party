#!/usr/bin/env python3
"""Sabotage-score api.go's PURE-function boundaries.

The sibling scorers each take one mechanism: sabotage-truncation.py and
sabotage-truncation-budgets.py the six cut sites, sabotage-naming-arithmetic.py
the naming and XP ladders, sabotage-journal-tables.py the journal and duration
tables, sabotage-store-paths.py the numbers behind a SQLite fixture, and
sabotage-request-boundaries.py the numbers a handler compares a REQUEST
against.

This one takes the corner of api.go that needs nothing at all: the gold-reward
arithmetic, the two tavern-banter thresholds, and the recent-tool-call feed,
whose data is declared mock. Every case runs against a bare `&Server{}` or a
plain function call.

  THE PRIOR COLUMN IS THE WHOLE REPO, NOT THE PACKAGE AND NOT THE FILE.

Inherited from the 205th pass, and it earns its keep here: two rows below
belong to spawn_message_budget_test.go, a live neighbour in this same package,
and are carried so they report PRIOR=CAUGHT -> "already held". If either ever
reads "gap closed", the base branch is missing that neighbour and every other
number in this table is inflated. The 208th's rule: score what you drop from
scope, do not delete it with a comment.

  THE BASE BRANCH IS AN INPUT TO THE SCORE.

The 207th's. Base on test/the-request-validated-boundaries-are-unpinned, which
contains all ten siblings and main. From main the two neighbour rows read as
gaps this pass closed.

  A DOMINATED MUTATION IS A DECLARED CONTROL, NOT AN OPEN GAP —
  AND THE THING DOMINATING IT NEED NOT BE IN THE SAME FUNCTION.

The 207th's rule met two new mechanisms here, and between them they account for
five of the six rows this sweep could not close:

  1. A CAP FURTHER DOWN THE CALL CHAIN. Both tool-call handlers validate a
     request limit and then hand it to getRecentToolCalls, whose loop is
     `i < limit/2 && i < 10`. The response length is therefore min(limit/2, 10),
     which is 10 for EVERY limit at or above 20. So the default 50 and the
     ceiling 100 have no straddle pair at all: accepting 100 and falling back to
     50 render the same ten rows. Enumerated over every limit from -5 to 204
     plus the empty and unparseable spellings, `<= 100` and `<= 99` agree at
     every input, as do `limit := 50` and `limit := 49`. The agent-scoped
     handler's default of 20 sits exactly at the knee (20/2 == 10), so it is
     observable when it drifts DOWN and not when it drifts up.

     Reading the request-validation block alone would have filed four gaps that
     no test can close. The 211th found dominance from a value being its own
     guard; this is dominance from a constant in a different function.

  2. A GAP IN THE ARITHMETIC ITSELF. `if goldReward < 5 { goldReward = 5 }`
     looks like the 211th's shape — one number spelled twice — but is worse:
     widening it to `< 6` would only move an input whose raw reward is exactly
     5, and 5 is unreachable. The base is `xpReward * 2`, always even, and the
     five multipliers take it to {2x, 1.6x, 3x, 4x, 6x}, whose values below 6
     are 0, 1, 2, 3 and 4. `< 7` IS observable, at raw 6. So the guard is
     scored by a TWO-step straddle and the one-step widening is a control.

  A LADDER SPELLED TWICE IS TWO LADDERS, AND THE SECOND MAY BE UNREACHABLE
  BY ANY CANONICAL INPUT.

calculateGoldReward spells 0.8/1.0/1.5/2.0/3.0 once over `case "easy", "1"` and
once over `case 1` inside the strconv.Atoi arm. The second copy is NOT the digit
path: "3" is taken by `case "hard", "3"` before Atoi is called. It is reached
only by a spelling Atoi accepts and the switch does not — "+3", "03", "005". So
five shipped literals sit behind inputs no ordinary caller sends, and a grep
over `multiplier = ` reports ten hits that a reader collapses into five. Every
one of the ten is scored separately below; the two ladders' rows are anchored on
their own `case` line, because the assignments are character-for-character
identical.

Rules inherited from earlier passes, each of which cost a session to learn:

  - The needle must occur EXACTLY once and the file's bytes must change.
  - A compile error is not a red. It is a separate verdict.
  - A panic in a source frame is detection; a panic in a _test.go frame is the
    fixture falling over and is not.
  - Controls in both directions, so a harness that reports CAUGHT for
    everything cannot look perfect.
  - Restores from git, so the tree must be committed before running.

Run:  python3 scripts/sabotage-purefunctions.py
      python3 scripts/sabotage-purefunctions.py --self-test
"""

import os
import pathlib
import re
import signal
import subprocess
import sys

REPO = pathlib.Path(__file__).resolve().parent.parent

API = "internal/api/api.go"

TRACKED = [API]

# The whole repo, for the reason in the docstring.
PACKAGES = ["./..."]

# The file this sweep adds. Renamed aside to build the PRIOR column.
MINE = ["internal/api/pure_function_boundaries_test.go"]

# Reach guards, by message. Inherited from the truncation tests, which share
# this package: a guard firing means a fixture stopped reaching the code, which
# is not the same as an assertion noticing a changed value.
GUARD_MARKERS = (
    'input no longer reaches the truncating branch',
    "no longer reaches generateQuestName's truncating fallback",
    'is missing from this working tree',
    'cannot tell which arm produced',
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
        ("--- FAIL: T\n    x_test.go:9: cannot tell which arm produced \"x\"\n", 1,
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


# --- Anchors ---------------------------------------------------------------
#
# `multiplier = 0.8` and its four siblings appear TWICE each, once per ladder,
# and the two assignments are identical down to the value. Only the case line
# above them differs, so every multiplier row carries its own case line.

NAMED_LADDER = {
    "0.8": 'case "easy", "1":\n\t\tmultiplier = 0.8',
    "1.0": 'case "medium", "2":\n\t\tmultiplier = 1.0',
    "1.5": 'case "hard", "3":\n\t\tmultiplier = 1.5',
    "2.0": 'case "expert", "4":\n\t\tmultiplier = 2.0',
    "3.0": 'case "legendary", "5":\n\t\tmultiplier = 3.0',
}

ATOI_LADDER = {
    "0.8": 'case 1:\n\t\t\t\tmultiplier = 0.8',
    "1.0": 'case 2:\n\t\t\t\tmultiplier = 1.0',
    "1.5": 'case 3:\n\t\t\t\tmultiplier = 1.5',
    "2.0": 'case 4:\n\t\t\t\tmultiplier = 2.0',
    "3.0": 'case 5:\n\t\t\t\tmultiplier = 3.0',
}

# The banter gate spells the same threshold twice on one line, once per agent,
# so a bare `> 2` needle would mutate whichever side comes first and report it
# under the other's name. Each row rewrites the whole line.
BANTER_GATE = "if agentA.RecentQuests > 2 || agentB.RecentQuests > 2 {"

# The two tool-call handlers' validation lines differ only in the ceiling, and
# their defaults differ only in the trailing comment. Both are unique as
# written; spelled out here so a reader can check the row names the site.
FEED_VALIDATION = "err == nil && parsedLimit > 0 && parsedLimit <= 100 {"
AGENT_FEED_VALIDATION = "err == nil && parsedLimit > 0 && parsedLimit <= 50 {"
FEED_DEFAULT = "limit := 50 // Default limit"
AGENT_FEED_DEFAULT = "limit := 20 // Default limit for single agent"

DERIVED_CAP = "for i := 0; i < limit/2 && i < 10; i++ {"


def ladder(table, value, replacement):
    """One multiplier row, anchored on its own case line."""
    find = table[value]
    return [(API, find, find.replace("= " + value, "= " + replacement, 1))]


# (label, [(file, find, replace)], site)
#
# Every mutation is a DRIFTED VALUE, not a deletion: a number is far likelier to
# drift by one than a whole check is to disappear, and a drift keeps every
# identifier live so the case reports a score instead of a compile error.
CASES = [
    # -- calculateGoldReward: the base conversion --------------------------
    ("the gold base conversion drifts up to 3 gold per XP",
     [(API, "baseGold := xpReward * 2", "baseGold := xpReward * 3")], "gold base 2"),

    ("the gold base conversion drifts down to 1 gold per XP",
     [(API, "baseGold := xpReward * 2", "baseGold := xpReward * 1")], "gold base 2"),

    # -- calculateGoldReward: the NAMED ladder -----------------------------
    ("the easy multiplier drifts to 0.9", ladder(NAMED_LADDER, "0.8", "0.9"),
     "named ladder easy 0.8"),
    ("the medium multiplier drifts to 1.1", ladder(NAMED_LADDER, "1.0", "1.1"),
     "named ladder medium 1.0"),
    ("the hard multiplier drifts to 1.6", ladder(NAMED_LADDER, "1.5", "1.6"),
     "named ladder hard 1.5"),
    ("the expert multiplier drifts to 2.1", ladder(NAMED_LADDER, "2.0", "2.1"),
     "named ladder expert 2.0"),
    ("the legendary multiplier drifts to 3.1", ladder(NAMED_LADDER, "3.0", "3.1"),
     "named ladder legendary 3.0"),

    # -- calculateGoldReward: the ATOI ladder ------------------------------
    # Five more literals, reachable only by "+3"/"03"/"005". A test that sends
    # only canonical difficulties leaves all five free while appearing to cover
    # the ladder, because the named rows above pass.
    ("the Atoi ladder's tier-1 multiplier drifts to 0.9", ladder(ATOI_LADDER, "0.8", "0.9"),
     "Atoi ladder 1 -> 0.8"),
    ("the Atoi ladder's tier-2 multiplier drifts to 1.1", ladder(ATOI_LADDER, "1.0", "1.1"),
     "Atoi ladder 2 -> 1.0"),
    ("the Atoi ladder's tier-3 multiplier drifts to 1.6", ladder(ATOI_LADDER, "1.5", "1.6"),
     "Atoi ladder 3 -> 1.5"),
    ("the Atoi ladder's tier-4 multiplier drifts to 2.1", ladder(ATOI_LADDER, "2.0", "2.1"),
     "Atoi ladder 4 -> 2.0"),
    ("the Atoi ladder's tier-5 multiplier drifts to 3.1", ladder(ATOI_LADDER, "3.0", "3.1"),
     "Atoi ladder 5 -> 3.0"),

    # -- calculateGoldReward: the minimum floor ----------------------------
    ("the gold floor GUARD tightens, so a reward of 4 is no longer lifted",
     [(API, "if goldReward < 5 {", "if goldReward < 4 {")], "gold floor guard"),

    # The one-step widening is DOMINATED and is scored as a control below.
    ("the gold floor GUARD widens by two, so a reward of 6 is lifted",
     [(API, "if goldReward < 5 {", "if goldReward < 7 {")], "gold floor guard"),

    ("the gold floor VALUE drifts up to 6",
     [(API, "\t\tgoldReward = 5", "\t\tgoldReward = 6")], "gold floor value"),

    ("the gold floor VALUE drifts down to 4",
     [(API, "\t\tgoldReward = 5", "\t\tgoldReward = 4")], "gold floor value"),

    # -- generateBanterMessage: the gate that opens the conversation -------
    ("the workload gate tightens on agentA, so three quests stay silent",
     [(API, BANTER_GATE,
       "if agentA.RecentQuests > 3 || agentB.RecentQuests > 2 {")], "banter gate agentA > 2"),

    ("the workload gate loosens on agentA, so two quests start a conversation",
     [(API, BANTER_GATE,
       "if agentA.RecentQuests > 1 || agentB.RecentQuests > 2 {")], "banter gate agentA > 2"),

    ("the workload gate tightens on agentB",
     [(API, BANTER_GATE,
       "if agentA.RecentQuests > 2 || agentB.RecentQuests > 3 {")], "banter gate agentB > 2"),

    ("the workload gate loosens on agentB",
     [(API, BANTER_GATE,
       "if agentA.RecentQuests > 2 || agentB.RecentQuests > 1 {")], "banter gate agentB > 2"),

    # -- getWorkloadMessage: the SECOND, higher pair -----------------------
    # A different number from the gate, in a different function, deciding not
    # whether the agents speak but which of three voices they use.
    ("the busy threshold tightens on agentA, so four quests read as steady",
     [(API, "if agentA.RecentQuests > 3 {", "if agentA.RecentQuests > 4 {")],
     "workload voice agentA > 3"),

    ("the busy threshold loosens on agentA, so three quests read as busy",
     [(API, "if agentA.RecentQuests > 3 {", "if agentA.RecentQuests > 2 {")],
     "workload voice agentA > 3"),

    ("the busy threshold tightens on agentB",
     [(API, "} else if agentB.RecentQuests > 3 {", "} else if agentB.RecentQuests > 4 {")],
     "workload voice agentB > 3"),

    ("the busy threshold loosens on agentB",
     [(API, "} else if agentB.RecentQuests > 3 {", "} else if agentB.RecentQuests > 2 {")],
     "workload voice agentB > 3"),

    # -- the recent-tool-call feed: the gate that IS separable --------------
    # Unlike the ceiling and the default beside it, `> 0` is observable at its
    # own boundary in both directions, because a rejected limit falls back to
    # ten rows and an accepted 1 renders none.
    ("the feed's positive-limit gate loosens, so zero is accepted",
     [(API, FEED_VALIDATION, FEED_VALIDATION.replace("parsedLimit > 0", "parsedLimit > -1"))],
     "feed gate > 0"),

    ("the feed's positive-limit gate tightens, so one is rejected",
     [(API, FEED_VALIDATION, FEED_VALIDATION.replace("parsedLimit > 0", "parsedLimit > 1"))],
     "feed gate > 0"),

    ("the agent feed's positive-limit gate loosens, so zero is accepted",
     [(API, AGENT_FEED_VALIDATION,
       AGENT_FEED_VALIDATION.replace("parsedLimit > 0", "parsedLimit > -1"))],
     "agent feed gate > 0"),

    ("the agent feed's positive-limit gate tightens, so one is rejected",
     [(API, AGENT_FEED_VALIDATION,
       AGENT_FEED_VALIDATION.replace("parsedLimit > 0", "parsedLimit > 1"))],
     "agent feed gate > 0"),

    # -- the ceilings, in the only direction that is observable at all ------
    # +-1 is dominated (see the controls). A ceiling that falls below 20 starts
    # rejecting limits whose exact value still shows in the response, and THAT
    # is what these rows pin. Filed as cases rather than controls because a
    # test can and does separate them.
    ("the feed ceiling collapses to 10, so a limit of 12 is rejected",
     [(API, FEED_VALIDATION, FEED_VALIDATION.replace("<= 100", "<= 10"))],
     "feed ceiling 100"),

    ("the agent feed ceiling collapses to 10",
     [(API, AGENT_FEED_VALIDATION, AGENT_FEED_VALIDATION.replace("<= 50", "<= 10"))],
     "agent feed ceiling 50"),

    # -- the defaults, likewise ---------------------------------------------
    ("the feed default collapses to 5",
     [(API, FEED_DEFAULT, "limit := 5 // Default limit")], "feed default 50"),

    ("the agent feed default drifts DOWN to 19, off the knee",
     [(API, AGENT_FEED_DEFAULT, "limit := 19 // Default limit for single agent")],
     "agent feed default 20"),

    # -- getRecentToolCalls: the derived cap --------------------------------
    # Both numbers live in a loop condition, so neither appears in a grep for
    # comparisons against a request field. This is the site the onboarding note
    # names as invisible to the grep that sized the card.
    ("the half-the-limit divisor drifts to a third",
     [(API, DERIVED_CAP, DERIVED_CAP.replace("limit/2", "limit/3"))], "feed divisor 2"),

    ("the limit is no longer halved at all",
     [(API, DERIVED_CAP, DERIVED_CAP.replace("limit/2", "limit"))], "feed divisor 2"),

    ("the row cap drifts down to 9",
     [(API, DERIVED_CAP, DERIVED_CAP.replace("i < 10", "i < 9"))], "feed row cap 10"),

    ("the row cap drifts up to 11",
     [(API, DERIVED_CAP, DERIVED_CAP.replace("i < 10", "i < 11"))], "feed row cap 10"),

    # -- getRecentToolCalls: the mock tables --------------------------------
    # Mock values, so nothing downstream depends on them — but they are the
    # part of this function a reader is likeliest to tidy, and the count
    # assertions above cannot see any of it.
    ("the tool-name cycle stops using the whole table",
     [(API, "tools[i%len(tools)]", "tools[i%3]")], "tool name cycle"),

    ("the synthetic agent cycle drifts to four names",
     [(API, 'fmt.Sprintf("agent_%d", i%3)', 'fmt.Sprintf("agent_%d", i%4)')],
     "agent name cycle 3"),

    ("a duration is attached every third row instead of every second",
     [(API, "if i%2 == 0 {", "if i%3 == 0 {")], "duration cadence 2"),

    ("the duration step drifts to 0.6 seconds",
     [(API, "duration := float64(i+1) * 0.5", "duration := float64(i+1) * 0.6")],
     "duration step 0.5"),

    # -- The spawn-message cut, scored to PROVE it is a neighbour's ---------
    # The 208th's rule. The card's own scope said to drop these two if the
    # truncation child was done; it is, and dropping them by reading would make
    # this table's number smaller with nothing to catch a misreading. As cases
    # they cost ~40 seconds and assert the base branch: if either reports
    # "gap closed", spawn_message_budget_test.go is missing from the base and
    # every row above is inflated.
    ("the spawn message cut's GUARD drifts up by one",
     [(API, "if len(task) > 100 {", "if len(task) > 101 {")],
     "spawn cut guard (neighbour)"),

    ("the spawn message cut's BUDGET drifts down by one",
     [(API, "TruncateAtRuneBoundary(task, 100)", "TruncateAtRuneBoundary(task, 99)")],
     "spawn cut budget (neighbour)"),
]

# (label, edits, expect_caught_by_mine, why)
CONTROLS = [
    ("KNOWN-POSITIVE: the gold floor collapses to 500",
     [(API, "\t\tgoldReward = 5", "\t\tgoldReward = 500")], True,
     "this sweep asserts four separate inputs are lifted to exactly 5, so a "
     "harness that cannot report CAUGHT here is dead"),

    ("KNOWN-POSITIVE: the tool-call row cap collapses to 2",
     [(API, DERIVED_CAP, DERIVED_CAP.replace("i < 10", "i < 2"))], True,
     "ten rows are asserted at four separate limits"),

    ("KNOWN-NEGATIVE: the gold floor guard widens by ONE, to `< 6`",
     [(API, "if goldReward < 5 {", "if goldReward < 6 {")], False,
     "DOMINATED by the arithmetic, not a fixture gap: this only moves an input "
     "whose raw reward is exactly 5, and no input produces 5. baseGold is "
     "xpReward*2, always even, and the multipliers 0.8/1.0/1.5/2.0/3.0 take it "
     "to {2x, 1.6x, 3x, 4x, 6x}, whose values under 6 are 0,1,2,3,4. The "
     "two-step widening to `< 7` IS a case above, and it is the only reason "
     "this guard is pinned in that direction"),

    ("KNOWN-NEGATIVE: the feed ceiling drifts down to `<= 99`",
     [(API, FEED_VALIDATION, FEED_VALIDATION.replace("<= 100", "<= 99"))], False,
     "DOMINATED by `i < 10` in getRecentToolCalls, a different function. The "
     "response length is min(limit/2, 10), so at limit=100 the accepted path "
     "renders 10 rows and the rejected path falls back to the default 50, "
     "which also renders 10. Enumerated over -5..204 plus the empty and "
     "unparseable spellings, the mutant answers identically at every input"),

    ("KNOWN-NEGATIVE: the feed ceiling drifts up to `<= 101`",
     [(API, FEED_VALIDATION, FEED_VALIDATION.replace("<= 100", "<= 101"))], False,
     "same domination, from the other side: 101 is accepted rather than "
     "rejected and renders the same ten rows"),

    ("KNOWN-NEGATIVE: the feed default drifts to 49",
     [(API, FEED_DEFAULT, "limit := 49 // Default limit")], False,
     "same cap, applied to the fallback: 49/2 and 50/2 both exceed 10. Any "
     "default at or above 20 is indistinguishable from any other, which is why "
     "the case above collapses it to 5 instead of drifting it by one"),

    ("KNOWN-NEGATIVE: the feed default drifts to 51",
     [(API, FEED_DEFAULT, "limit := 51 // Default limit")], False,
     "same reason, upward"),

    ("KNOWN-NEGATIVE: the agent feed default drifts UP to 21",
     [(API, AGENT_FEED_DEFAULT, "limit := 21 // Default limit for single agent")], False,
     "20 sits exactly at the knee: 20/2 == 10 == the cap, so drifting UP is "
     "hidden and drifting DOWN is not. The downward drift to 19 is a case "
     "above and it is CAUGHT. This asymmetry is the whole difference between "
     "the two handlers"),

    ("KNOWN-NEGATIVE: the agent feed ceiling drifts to `<= 49`",
     [(API, AGENT_FEED_VALIDATION, AGENT_FEED_VALIDATION.replace("<= 50", "<= 49"))], False,
     "the same domination as its sibling's ceiling, one handler over"),
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
        # mutate a site the row does not name — which is not hypothetical
        # here, where each multiplier is spelled twice.
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
