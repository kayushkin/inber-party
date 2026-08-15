#!/usr/bin/env python3
"""Sabotage-score inber.go's SQLite STORE-PATH numbers.

The sibling scorers cover this file's pure functions: sabotage-truncation.py and
sabotage-truncation-budgets.py score the six cut sites, sabotage-naming-
arithmetic.py scores the naming and XP arithmetic, sabotage-journal-tables.py
scores the journal and duration tables. All of those reach their targets by
calling a function directly.

This one scores the numbers that only appear once a *database* is open: the
quest difficulty ladder, the quest progress constants, the `limit <= 0`
substituted defaults, the GetAgents status heuristics, and the two tool-call
skill gates. Every case here drives a real SQLite file built in a t.TempDir().

  THE PRIOR COLUMN IS THE WHOLE REPO, NOT THE PACKAGE AND NOT THE FILE.

Inherited from the 205th pass and re-confirmed by the 207th, which found 9 of 24
rows already held once the baseline was widened. This sweep has a live sibling
risk of its own: truncation_budgets_test.go drives (*Store).GetQuests through
the same gateway fixture to reach the 200-byte description cut, so it is one
assertion away from several rows below. PRIOR renames this sweep's file aside
and runs `go test ./...`; anything a neighbour already holds reports
PRIOR=CAUGHT and lands in "already covered" instead of the closed-gaps count.

  THE BASE BRANCH IS AN INPUT TO THE SCORE.

Also the 207th's, and it bites harder here than anywhere. This branch is based
on the merge of eight sibling branches, and one of them
(test/the-truncation-budgets-are-unpinned) already pins the 200-byte quest
description cut that this sweep's own card listed as in scope. Run this from
`main` and that row reads as a gap this pass closed. It is not. Name your base
branch in the write-up or the number cannot be checked.

  AN UNREACHABLE MUTATION IS A DECLARED CONTROL, NOT AN OPEN GAP.

GetConversations spells the same `limit <= 0 -> 50` substitution as GetQuests,
but its query selects sessions.parent_session_id, initial_message and
last_message_at, none of which any fixture in this package defines. The query
errors before the limit can affect a returned row, so no test can catch the
mutation without a schema change. It is scored as a known-NEGATIVE beside the
harness's own controls rather than filed as a gap, because filing it would send
the next pass after a test that cannot exist at today's fixture.

Rules inherited from earlier passes, each of which cost a session to learn:

  - The needle must occur EXACTLY once and the file's bytes must change. Three
    of this sweep's targets are spelled twice in inber.go — `limit = 50` (in
    GetQuests and in GetConversations) and both `if toolCalls > 0 {` /
    `levelForXP(toolCalls * 10)` (the agent-merge branch and the sessions-only
    branch) — so each is matched with enough surrounding lines to name one site.
    A bare needle would silently mutate the first of the pair and score it under
    the other's name.
  - A compile error is not a red. It is a separate verdict.
  - A panic in a source frame is detection; a panic in a _test.go frame is the
    fixture falling over and is not.
  - Two controls, a known-positive and a known-negative, so a harness that
    reports CAUGHT for everything cannot look perfect.
  - Restores from git, so the tree must be committed before running.

Run:  python3 scripts/sabotage-store-paths.py
      python3 scripts/sabotage-store-paths.py --self-test
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

# The whole repo, for the reason in the docstring.
PACKAGES = ["./..."]

# The file this sweep adds. Renamed aside to build the PRIOR column.
MINE = ["internal/inber/store_path_boundaries_test.go"]

# Reach guards, by message. Inherited from the truncation tests, which share
# these packages. This sweep contributes none: every test here asserts a value
# rather than guarding that an input still reaches a branch.
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


# --- Anchors for the three targets inber.go spells twice --------------------
#
# Written out here rather than inline so that a reader can check the anchor
# names the site the row claims. Getting one wrong mutates the other site and
# reports the result under this row's name, which is the failure the
# exactly-once rule exists to prevent and which an anchor can still slip past.

# `limit = 50` is in GetQuests and in GetConversations. Anchor on the signature.
QUESTS_LIMIT = ('func (s *Store) GetQuests(limit int) ([]RPGQuest, error) {\n'
                '\tif limit <= 0 {\n'
                '\t\tlimit = 50')
CONVERSATIONS_LIMIT = ('func (s *Store) GetConversations(limit int) ([]RPGConversation, error) {\n'
                       '\tif limit <= 0 {\n'
                       '\t\tlimit = 50')
HISTORY_LIMIT = ('func (s *Store) GetQuestHistory(agentID string, limit int) '
                 '([]QuestHistoryEntry, error) {\n'
                 '\tif limit <= 0 {\n'
                 '\t\tlimit = 20')

# The tool-call skill block is byte-identical in both GetAgents branches. The
# first is reached when the agent is already in agentMap from the gateway DB;
# the second when the sessions DB is its only source. Anchor the first on the
# `agentMap` lookup above it and the second on the assignment below it.
SKILL_BODY = ('''\t\t\t\t\t\ttoolLevel, _ := levelForXP(toolCalls * 10)
\t\t\t\t\t\ta.Skills = append(a.Skills, RPGSkill{
\t\t\t\t\t\t\tName:      "Tool Mastery",
\t\t\t\t\t\t\tLevel:     toolLevel,
\t\t\t\t\t\t\tTaskCount: toolCalls,
\t\t\t\t\t\t})''')

MERGE_SITE = ('\t\t\t\tif a, ok := agentMap[agentName]; ok {\n'
              '\t\t\t\t\tif toolCalls > 0 {\n' + SKILL_BODY)
SESSIONS_ONLY_SITE = ('\t\t\t\t\tif toolCalls > 0 {\n' + SKILL_BODY + '\n'
                      '\t\t\t\t\t}\n'
                      '\t\t\t\t\tagentMap[agentName] = a')

# (label, [(file, find, replace)], site)
#
# Every mutation is a DRIFTED VALUE, not a deletion: a number is far likelier to
# drift by one than a whole check is to disappear, and a drift keeps every
# identifier live so the case reports a score instead of a compile error.
CASES = [
    # -- GetQuests: the four difficulty thresholds, both directions ---------
    # Each threshold is strictly-greater, so it is pinned only by the pair
    # (T, T+1). Scoring both directions is what shows a one-sided pin.
    ("the top difficulty threshold drifts up, so 5001 tokens stay at tier 4",
     [(INBER, "if totalTokens > 5000 {", "if totalTokens > 5001 {")],
     "difficulty threshold 5000"),

    ("the top difficulty threshold drifts down, so 5000 tokens reach tier 5",
     [(INBER, "if totalTokens > 5000 {", "if totalTokens > 4999 {")],
     "difficulty threshold 5000"),

    ("the tier-4 threshold drifts up, so 2001 tokens stay at tier 3",
     [(INBER, "} else if totalTokens > 2000 {", "} else if totalTokens > 2001 {")],
     "difficulty threshold 2000"),

    ("the tier-4 threshold drifts down, so 2000 tokens reach tier 4",
     [(INBER, "} else if totalTokens > 2000 {", "} else if totalTokens > 1999 {")],
     "difficulty threshold 2000"),

    ("the tier-3 threshold drifts up, so 1001 tokens stay at tier 2",
     [(INBER, "} else if totalTokens > 1000 {", "} else if totalTokens > 1001 {")],
     "difficulty threshold 1000"),

    ("the tier-3 threshold drifts down, so 1000 tokens reach tier 3",
     [(INBER, "} else if totalTokens > 1000 {", "} else if totalTokens > 999 {")],
     "difficulty threshold 1000"),

    ("the tier-2 threshold drifts up, so 501 tokens stay at tier 1",
     [(INBER, "} else if totalTokens > 500 {", "} else if totalTokens > 501 {")],
     "difficulty threshold 500"),

    ("the tier-2 threshold drifts down, so 500 tokens reach tier 2",
     [(INBER, "} else if totalTokens > 500 {", "} else if totalTokens > 499 {")],
     "difficulty threshold 500"),

    # -- GetQuests: the tier VALUES, which the thresholds cannot pin --------
    # A ladder returning 6/4/3/2/1 crosses at exactly the same token counts as
    # one returning 5/4/3/2/1, so only a test naming the value separates them.
    ("the base difficulty drifts up, so a zero-token quest is tier 2",
     [(INBER, "\t\t\tdifficulty := 1", "\t\t\tdifficulty := 2")],
     "difficulty value 1"),

    ("the top tier drifts up to 6",
     [(INBER, "difficulty = 5", "difficulty = 6")], "difficulty value 5"),

    ("tier 4 drifts down to 3",
     [(INBER, "difficulty = 4", "difficulty = 3")], "difficulty value 4"),

    ("tier 3 drifts up to 4",
     [(INBER, "difficulty = 3", "difficulty = 4")], "difficulty value 3"),

    ("tier 2 drifts down to 1",
     [(INBER, "difficulty = 2", "difficulty = 1")], "difficulty value 2"),

    # -- GetQuests: the four progress constants -----------------------------
    # Assigned outputs selected by status, with no ordering for a test to lean
    # on. TestGetQuests asserts count, status mix and Children, never Progress.
    ("the in-progress default drifts up by one",
     [(INBER, "progress := 50", "progress := 51")], "progress default 50"),

    ("the completed progress drifts down, so a finished quest reads 99",
     [(INBER, "progress = 100", "progress = 99")], "progress completed 100"),

    ("the failed progress drifts up by one",
     [(INBER, "progress = 30 // died trying", "progress = 31 // died trying")],
     "progress failed 30"),

    ("the available progress drifts up, so an unstarted quest reads 1",
     [(INBER, "progress = 0", "progress = 1")], "progress available 0"),

    # -- The 200-byte description cut, scored to PROVE it is already held ---
    # The card listed this cut as in scope. It was done one pass earlier, by
    # truncation_budgets_test.go on test/the-truncation-budgets-are-unpinned,
    # which is in this branch's base. Dropping it from the case table would
    # have made that a claim; scoring it makes it a measurement, and the two
    # rows below are expected to report PRIOR=CAUGHT -> "already held" rather
    # than to count toward this sweep's closed gaps. If they ever read
    # "gap closed", the base branch is wrong and every number here is inflated.
    ("the description cut's GUARD drifts up by one",
     [(INBER, "if len(questDesc) > 200 {", "if len(questDesc) > 201 {")],
     "description cut guard (neighbour)"),

    ("the description cut's BUDGET drifts down by one",
     [(INBER, "TruncateAtRuneBoundary(questDesc, 200)",
       "TruncateAtRuneBoundary(questDesc, 199)")],
     "description cut budget (neighbour)"),

    # -- The substituted `limit <= 0` defaults ------------------------------
    # The 184th's rule: a parameter every test supplies is not the value that
    # ships. Both callers below can pass nothing, and neither default was
    # entered by any test.
    ("GetQuests' substituted limit drifts down to 49",
     [(INBER, QUESTS_LIMIT, QUESTS_LIMIT.replace("limit = 50", "limit = 49"))],
     "GetQuests default limit"),

    ("GetQuests' substituted limit drifts up to 51",
     [(INBER, QUESTS_LIMIT, QUESTS_LIMIT.replace("limit = 50", "limit = 51"))],
     "GetQuests default limit"),

    ("GetQuestHistory's substituted limit drifts down to 19",
     [(INBER, HISTORY_LIMIT, HISTORY_LIMIT.replace("limit = 20", "limit = 19"))],
     "GetQuestHistory default limit"),

    ("GetQuestHistory's substituted limit drifts up to 21",
     [(INBER, HISTORY_LIMIT, HISTORY_LIMIT.replace("limit = 20", "limit = 21"))],
     "GetQuestHistory default limit"),

    # -- GetAgents: the status heuristics -----------------------------------
    # TestGetAgents_GatewayOnly asserts Class, Tokens, ErrorCount and Name and
    # never reads Status, so the whole heuristic was free.
    ("the running gate tightens, so a single running request is not working",
     [(INBER, "if running > 0 {", "if running > 1 {")], "running gate"),

    ("the error gate tightens, so a lone errored request is not stuck",
     [(INBER, "errors > 0 &&", "errors > 1 &&")], "error gate"),

    ("the stuck ratio rises to 0.6, so three errors in five are not stuck",
     [(INBER, "> 0.5 {", "> 0.6 {")], "error ratio 0.5"),

    ("the stuck ratio falls to 0.4, so two errors in five are stuck",
     [(INBER, "> 0.5 {", "> 0.4 {")], "error ratio 0.5"),

    ("the idle status string drifts",
     [(INBER, 'status := "idle"', 'status := "resting"')], "status value idle"),

    ("the working status string drifts",
     [(INBER, 'status = "working"', 'status = "busy"')], "status value working"),

    ("the stuck status string drifts",
     [(INBER, 'status = "stuck"', 'status = "blocked"')], "status value stuck"),

    # -- GetAgents: the two tool-call skill gates ---------------------------
    # The same rule in two branches. A test exercising one leaves the other
    # free, which is why both are scored and why the tests drive both in one
    # call via an agent in both databases and an agent in only the sessions DB.
    ("the merge branch's tool gate tightens, so one tool call earns no skill",
     [(INBER, MERGE_SITE, MERGE_SITE.replace("if toolCalls > 0 {", "if toolCalls > 1 {"))],
     "tool gate (merge branch)"),

    ("the sessions-only branch's tool gate tightens",
     [(INBER, SESSIONS_ONLY_SITE,
       SESSIONS_ONLY_SITE.replace("if toolCalls > 0 {", "if toolCalls > 1 {"))],
     "tool gate (sessions-only)"),

    # The `* 10` is a call ARGUMENT, so no grep over comparison lines sees it —
    # the 204th's point about api.go's TruncateAtRuneBoundary(task, 100), and it
    # recurs here twice. It is the number that decides the rendered skill level.
    ("the merge branch's XP-per-tool-call drifts up to 11",
     [(INBER, MERGE_SITE, MERGE_SITE.replace("toolCalls * 10", "toolCalls * 11"))],
     "tool XP multiplier (merge)"),

    ("the merge branch's XP-per-tool-call drifts down to 9",
     [(INBER, MERGE_SITE, MERGE_SITE.replace("toolCalls * 10", "toolCalls * 9"))],
     "tool XP multiplier (merge)"),

    ("the sessions-only branch's XP-per-tool-call drifts up to 11",
     [(INBER, SESSIONS_ONLY_SITE, SESSIONS_ONLY_SITE.replace("toolCalls * 10", "toolCalls * 11"))],
     "tool XP multiplier (sessions)"),

    ("the sessions-only branch's XP-per-tool-call drifts down to 9",
     [(INBER, SESSIONS_ONLY_SITE, SESSIONS_ONLY_SITE.replace("toolCalls * 10", "toolCalls * 9"))],
     "tool XP multiplier (sessions)"),
]

# (label, edits, expect_caught_by_mine, why)
CONTROLS = [
    ("KNOWN-POSITIVE: every completed quest is reported as failed",
     [(INBER, '\t\t\tcase "completed", "success":\n\t\t\t\tquestStatus = "completed"',
       '\t\t\tcase "completed", "success":\n\t\t\t\tquestStatus = "failed"')], True,
     "the status mix is asserted by TestGetQuests and by this sweep, so a "
     "harness that cannot report CAUGHT here is dead"),

    ("KNOWN-POSITIVE: the difficulty ladder is bypassed and every quest is tier 1",
     [(INBER, "if totalTokens > 5000 {", "if totalTokens > 99999999 {")], True,
     "collapses the top tier into tier 4; this sweep asserts tier 5 twice"),

    ("KNOWN-NEGATIVE: GetConversations' substituted limit drifts to 49",
     [(INBER, CONVERSATIONS_LIMIT, CONVERSATIONS_LIMIT.replace("limit = 50", "limit = 49"))],
     False,
     "unreachable at today's fixture: the query needs sessions.parent_session_id, "
     "initial_message and last_message_at, none of which any fixture defines, so "
     "it errors before the limit can affect a row. A schema change, not a gap"),

    ("KNOWN-NEGATIVE: the quest-name length cut drifts",
     [(INBER, "if len(questName) > 80 {", "if len(questName) > 81 {")], False,
     "not this sweep's target — it belongs to the truncation family and is "
     "already held by truncation_budgets_test.go, so this sweep must NOT claim "
     "it. Scored to prove the file aside is really this sweep's file"),
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
