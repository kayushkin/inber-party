#!/usr/bin/env python3
"""Sabotage-score the truncation BUDGET and GUARD values.

scripts/sabotage-truncation.py scores the rune-boundary *mechanism*: that the
cut never splits a rune. This scores the *numbers* — the guard that decides
whether to cut and the budget handed to the cut. They are a different question
and the mechanism scorer answers neither: every one of its cases replaces a
rune-safe cut with a byte cut and leaves both numbers alone.

Each of the six sites spells its budget twice, so each gets two rows. A budget
spelled twice is two boundaries; scoring them together hides a disagreement
between them.

  TWO COLUMNS, and the PRIOR one is built the expensive way on purpose.

PRIOR is the whole package with this sweep's new test files renamed out of the
tree — not a diff against the one test file this sweep looks like it extends. A
single-file baseline cannot see that a *sibling* test already covers the
mutation, so it reports a closed gap that was never open and flatters the score.
Anything a neighbour already holds reports PRIOR=CAUGHT here and lands in
"already covered" rather than in the closed-gaps count.

Rules inherited from earlier passes, each of which cost a session to learn:

  - The needle must occur EXACTLY once and the file's bytes must change. A
    replacement that matched nothing scores UNNOTICED for free and looks
    exactly like a gap in the tests. A replacement that matched twice mutates a
    site the row does not name.
  - A compile error is not a red. `go test` reports "[build failed]" with a
    non-zero exit code, which is indistinguishable from a caught mutation if
    you only read the exit code. It is a separate verdict.
  - A reach guard firing is not detection. This suite's guards are named in
    GUARD_MARKERS; a run that goes red only because the fixture stopped
    reaching the code under test has measured nothing.
  - Two controls. A known-positive every test must catch and a known-NEGATIVE
    no test should catch, so a harness that reports CAUGHT for everything
    cannot look perfect.
  - Restores from git, so the tree must be committed before running. A SIGTERM
    or SIGHUP between the write and the restore would otherwise leave a mutated
    source file behind looking like ordinary uncommitted work.

Run:  python3 scripts/sabotage-truncation-budgets.py
"""

import os
import pathlib
import re
import signal
import subprocess
import sys

REPO = pathlib.Path(__file__).resolve().parent.parent

TEXTUTIL = "internal/textutil/textutil.go"
INBER = "internal/inber/inber.go"
HTTPCLIENT = "internal/inber/http_client.go"
LOGSTACK = "internal/logstack/logstack.go"
API = "internal/api/api.go"

TRACKED = [TEXTUTIL, INBER, HTTPCLIENT, LOGSTACK, API]

PACKAGES = ["./internal/textutil/", "./internal/inber/",
            "./internal/logstack/", "./internal/api/"]

# The files this sweep adds. Renamed aside to build the PRIOR column.
MINE = [
    "internal/textutil/budget_boundary_test.go",
    "internal/inber/truncation_budgets_test.go",
    "internal/logstack/title_budget_test.go",
    "internal/api/spawn_message_budget_test.go",
]

# Reach guards, by message. Only this repo knows which of its own t.Fatalf
# lines are guards rather than assertions.
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
    # An assertion message is read before the stack: a run can carry both, and a
    # suite that asserts the defect in one test and crashes on it in another has
    # detected it either way.
    if real:
        return "CAUGHT"
    if "panic:" in output:
        # A mutation the test drove into a crash IS detection. A panic in the
        # fixture is the test falling over before asserting, which is not — and
        # the two are told apart only by which file the deepest non-runtime
        # frame names.
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
    """Whether a verdict means an assertion actually looked at the mechanism."""
    return verdict == "CAUGHT" or verdict.startswith("CAUGHT (panic in ")


def self_test():
    """No case table can score its own scorer, so drive every verdict."""
    probes = [
        ("--- FAIL: T\n    x_test.go:9: got 1 want 2\n", 1, "CAUGHT"),
        ("--- FAIL: T\n    x_test.go:9: a 60-byte input no longer reaches "
         "generateQuestName's truncating fallback, so this test proves nothing\n", 1,
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
    # A build failure must never be read as coverage — the trap this exists for.
    print("self-test: %s" % ("all verdicts reachable and separated" if ok else "FAILED"))
    return 0 if ok else 1


if "--self-test" in sys.argv:
    sys.exit(self_test())


# (label, [(file, find, replace)], site)
#
# Every mutation is a DRIFTED VALUE, not a deletion: a number is far likelier to
# drift by one than a whole check is to disappear, and a drift keeps every
# identifier live so the case reports a score instead of a compile error.
CASES = [
    # -- textutil: the helper's own two guards ------------------------------
    ("the empty-budget guard widens, so a budget of 1 yields nothing",
     [(TEXTUTIL, "if maxBytes <= 0 {", "if maxBytes <= 1 {")],
     "textutil guard"),

    ("the pass-through guard narrows, so a string of exactly maxBytes is cut",
     [(TEXTUTIL, "if len(s) <= maxBytes {", "if len(s) < maxBytes {")],
     "textutil guard"),

    # -- generateQuestName: guard 60, budget 57 -----------------------------
    ("the quest-name guard drifts up, so a 61-byte text is left whole",
     [(INBER, "if len(text) > 60 {", "if len(text) > 61 {")],
     "generateQuestName guard"),

    ("the quest-name guard drifts down, so a 60-byte text is cut",
     [(INBER, "if len(text) > 60 {", "if len(text) > 59 {")],
     "generateQuestName guard"),

    ("the quest-name budget drifts down by one byte",
     [(INBER, "textutil.TruncateAtRuneBoundary(text, 57)",
       "textutil.TruncateAtRuneBoundary(text, 56)")],
     "generateQuestName budget"),

    # -- truncateText: guard maxLen, budget maxLen-3 ------------------------
    ("truncateText's budget drifts, so it keeps one byte less than it says",
     [(INBER, "textutil.TruncateAtRuneBoundary(text, maxLen-3)",
       "textutil.TruncateAtRuneBoundary(text, maxLen-4)")],
     "truncateText budget"),

    ("truncateText's guard narrows, so text of exactly maxLen is cut",
     [(INBER, "if len(text) <= maxLen {", "if len(text) < maxLen {")],
     "truncateText guard"),

    # -- HTTPClient.GetQuests: guard 200, budget 200 ------------------------
    ("the HTTP description guard drifts up, so a 201-byte input is left whole",
     [(HTTPCLIENT, "if len(questDesc) > 200 {", "if len(questDesc) > 201 {")],
     "http_client guard"),

    ("the HTTP description budget drifts down by one byte",
     [(HTTPCLIENT, "textutil.TruncateAtRuneBoundary(questDesc, 200)",
       "textutil.TruncateAtRuneBoundary(questDesc, 199)")],
     "http_client budget"),

    # -- (*Store).GetQuests: the same cut against SQLite --------------------
    ("the store description guard drifts up, so a 201-byte input is left whole",
     [(INBER, "if len(questDesc) > 200 {", "if len(questDesc) > 201 {")],
     "store GetQuests guard"),

    ("the store description budget drifts down by one byte",
     [(INBER, "textutil.TruncateAtRuneBoundary(questDesc, 200)",
       "textutil.TruncateAtRuneBoundary(questDesc, 199)")],
     "store GetQuests budget"),

    # -- logstack conversation title: guard 60, budget 57 -------------------
    ("the title guard drifts up, so a 61-byte title is left whole",
     [(LOGSTACK, "if len(title) > 60 {", "if len(title) > 61 {")],
     "logstack guard"),

    ("the title budget drifts down by one byte",
     [(LOGSTACK, "textutil.TruncateAtRuneBoundary(title, 57)",
       "textutil.TruncateAtRuneBoundary(title, 56)")],
     "logstack budget"),

    # -- api spawn message: guard 100, budget 100 ---------------------------
    ("the spawn-task guard drifts up, so a 101-byte task is left whole",
     [(API, "if len(task) > 100 {", "if len(task) > 101 {")],
     "api guard"),

    # The budget the grep cannot see: it is a call ARGUMENT, so it appears in
    # no comparison in that file and a literal census misses it entirely.
    ("the spawn-task budget drifts down by one byte",
     [(API, "textutil.TruncateAtRuneBoundary(task, 100)",
       "textutil.TruncateAtRuneBoundary(task, 99)")],
     "api budget"),
]

# (label, edits, expect_caught_by_mine, why)
CONTROLS = [
    ("KNOWN-POSITIVE: the helper trims to nothing instead of to the boundary",
     [(TEXTUTIL, "return s[:cut]", "return s[:0]")], True,
     "every call site's assertion must notice its text vanishing"),

    ("KNOWN-NEGATIVE: the empty-budget guard narrows from <=0 to <0",
     [(TEXTUTIL, "if maxBytes <= 0 {", "if maxBytes < 0 {")], False,
     "a behavioural no-op: maxBytes==0 already yields \"\" through the cut below"),
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
            return "pattern occurs %d times in %s (want exactly 1): %r" % (n, fname, find[:60])
        path.write_text(text.replace(find, replace, 1))
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
