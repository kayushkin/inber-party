#!/usr/bin/env python3
"""Sabotage-score the held-item, tool-duration and journal-prose boundaries.

A passing suite is not evidence. This moves one number at a time and records
whether the suite notices, in TWO columns:

    PRIOR   the package as it was, with journal_boundaries_test.go removed
    MINE    the package with that file present

PRIOR is the whole package, not one test file. A single-file baseline cannot
see a sibling test that already covers the mutation, so it always flatters the
new file — measured elsewhere on this box as a "closed coverage gap" that a
neighbouring test had held all along.

Rules carried here, each inherited from a pass that lost time to its absence:

  - The needle must be unique and must actually change the file. A replacement
    matching nothing scores UNNOTICED for free, which is indistinguishable from
    a real gap. Every edit asserts exactly one occurrence.
  - A compile error is not a red. `go test` reports `[build failed]` with a
    non-zero exit, which looks exactly like a caught mutation if you only read
    the exit code. It gets its own verdict.
  - Two controls. A known-POSITIVE that both columns must catch, proving the
    harness can report CAUGHT at all, and a known-NEGATIVE that neither may
    catch, proving it is not simply reporting CAUGHT for everything.
  - Restores from git, so the tree must be committed before running.
  - Prints the edit it actually applied, because a row prints the label you
    gave it, not the change you made.

Run:  python3 scripts/sabotage-journal-tables.py
      python3 scripts/sabotage-journal-tables.py --self-test
"""

import os
import pathlib
import re
import signal
import subprocess
import sys

REPO = pathlib.Path(__file__).resolve().parent.parent

INBER = "internal/inber/inber.go"
HTTPCLIENT = "internal/inber/http_client.go"
PACKAGE = "./internal/inber/"

# Every file a case may mutate. Restored between cases and checked for
# uncommitted work before the run.
TRACKED = [INBER, HTTPCLIENT]

# The file under measurement. PRIOR runs with it moved aside.
MINE = REPO / "internal/inber/journal_boundaries_test.go"
MINE_PARKED = REPO / "internal/inber/journal_boundaries_test.go.parked"

# This suite has no fixture-reach guards — every t.Fatalf in the new file is an
# assertion about the mechanism, not a check that the input still reaches it.
# The distinction is kept because it is easy to add a guard later and forget
# that a guard firing is not coverage.
GUARD_MARKERS = ()

_FAIL_LINE = re.compile(r"^\s*(\S+_test\.go):(\d+): (.*)$", re.M)
_FRAME = re.compile(r"^\s+(/\S+\.go):(\d+)", re.M)


def counts_as_coverage(verdict):
    """Whether a verdict means an assertion actually looked at the mechanism."""
    return verdict == "CAUGHT" or verdict.startswith("CAUGHT (panic in ")


def classify(output, returncode):
    """Turn one `go test` run into a verdict.

    Build failures are separated from reds first: `go test` exits non-zero for
    both, and counting a syntax error as a caught mutation would report
    coverage that does not exist.
    """
    if "[build failed]" in output or "build failed" in output or "syntax error" in output:
        return "COMPILE ERROR", ""
    if returncode == 0:
        return "UNNOTICED", ""

    messages = [m for _, _, m in _FAIL_LINE.findall(output)]
    guard = [m for m in messages if any(g in m for g in GUARD_MARKERS)]
    real = [m for m in messages if m not in guard]
    if real:
        return "CAUGHT", real[0][:90]
    if "panic:" in output:
        frames = [f for f, _ in _FRAME.findall(output.split("panic:", 1)[1])
                  if f.startswith(str(REPO) + "/")]
        source = next((f for f in frames if not f.endswith("_test.go")), None)
        if source:
            return ("CAUGHT (panic in %s)" % pathlib.Path(source).name,
                    "the test drove the mutation into a crash")
        return "CAUGHT (fixture panicked)", "the test fell over before asserting"
    if guard:
        return "CAUGHT (guard)", guard[0][:90]
    return "CAUGHT (no message)", ""


def self_test():
    """Drive classify() through every verdict it can return.

    A scorer is the next unmeasured claim the moment you write one, and no case
    table can score its own scorer.
    """
    probes = [
        ("", 0, "UNNOTICED"),
        ("--- FAIL: T\n    x_test.go:9: Priority = 14, want 15\n", 1, "CAUGHT"),
        ("# pkg\n./inber.go:12:2: syntax error: unexpected }\nFAIL pkg [build failed]\n", 2,
         "COMPILE ERROR"),
        ("panic: boom\n\t/usr/lib/go/src/runtime/panic.go:8\n\t%s/inber.go:3\n" % REPO, 2,
         "CAUGHT (panic in inber.go)"),
        ("panic: boom\n\t%s/x_test.go:3\n" % REPO, 2, "CAUGHT (fixture panicked)"),
        ("--- FAIL: T\nno recognisable line\n", 1, "CAUGHT (no message)"),
    ]
    ok = True
    for output, rc, want in probes:
        got, _ = classify(output, rc)
        if got != want:
            print("SELF-TEST FAIL: classify(rc=%d) -> %r, want %r" % (rc, got, want))
            ok = False
    for verdict, want in [("CAUGHT", True), ("CAUGHT (panic in inber.go)", True),
                          ("CAUGHT (guard)", False), ("CAUGHT (fixture panicked)", False),
                          ("CAUGHT (no message)", False), ("UNNOTICED", False),
                          ("COMPILE ERROR", False)]:
        if counts_as_coverage(verdict) != want:
            print("SELF-TEST FAIL: counts_as_coverage(%r) != %r" % (verdict, want))
            ok = False
    print("self-test: %s" % ("all verdicts reachable and separated" if ok else "FAILED"))
    return 0 if ok else 1


if "--self-test" in sys.argv:
    sys.exit(self_test())

# (label, file, find, replace, expect_mine_catches, note)
#
# Every mutation is a DRIFTED VALUE or a FLIPPED COMPARISON rather than a
# deletion: values drift far more often than checks get removed, and a deletion
# tends to orphan an identifier and report a compile error instead of a score.
CASES = [
    # ---- getHeldItemsForAgent: activity thresholds ------------------------
    ("the spawn threshold drifts 2 -> 3",
     INBER,      "threshold = 2 // Lower threshold for spawning",
     "threshold = 3 // Lower threshold for spawning", True, ""),

    ("the spawn threshold drifts 2 -> 1",
     INBER,      "threshold = 2 // Lower threshold for spawning",
     "threshold = 1 // Lower threshold for spawning", True, ""),

    ("the infra threshold drifts 2 -> 3",
     INBER,      "threshold = 2 // Lower threshold for infra work",
     "threshold = 3 // Lower threshold for infra work", True, ""),

    ("the default threshold drifts 5 -> 6",
     INBER,      "threshold = 5 // Default threshold",
     "threshold = 6 // Default threshold", True, ""),

    ("the default threshold drifts 5 -> 4",
     INBER,      "threshold = 5 // Default threshold",
     "threshold = 4 // Default threshold", True, ""),

    ("the threshold comparison narrows from >= to >",
     INBER,      "if count >= threshold {", "if count > threshold {", True,
     "the boundary value itself stops qualifying"),

    # ---- getHeldItemsForAgent: the cap and the two sorts -------------------
    ("the held-item cap drifts 2 -> 3",
     INBER,      "maxItems := 2", "maxItems := 3", True, ""),

    ("the held-item cap drifts 2 -> 1",
     INBER,      "maxItems := 2", "maxItems := 1", True, ""),

    ("the priority sort reverses, so the lowest priority comes first",
     INBER,      "if heldItems[j].Priority > heldItems[i].Priority {",
     "if heldItems[j].Priority < heldItems[i].Priority {", True, ""),

    ("the score ranking reverses, so the cap keeps the least active",
     INBER,      "if scores[j].score > scores[i].score {",
     "if scores[j].score < scores[i].score {", True, ""),

    # ---- getHeldItemsForAgent: a catalog value ----------------------------
    ("the Claxon Horn's priority drifts 15 -> 14",
     INBER,      "Priority:     15,", "Priority:     14,", True,
     "still the highest, so only an absolute assertion sees this"),

    # ---- estimateToolDuration --------------------------------------------
    ("the shell-command estimate drifts 3.0 -> 3.5",
     INBER,      "return 3.0 // Shell commands take longer",
     "return 3.5 // Shell commands take longer", True, ""),

    ("the default estimate drifts 1.5 -> 1.0",
     INBER,      "return 1.5 // Default duration",
     "return 1.0 // Default duration", True,
     "collapses the default arm onto the file-operation arm"),

    # ---- generateActivityDescription: the token ladder --------------------
    ("the activity-description epic boundary drifts 1000 -> 1100",
     INBER,      "if tokens > 1000 {", "if tokens > 1100 {", True, ""),

    ("the activity-description epic boundary loosens > to >=",
     INBER,      "if tokens > 1000 {", "if tokens >= 1000 {", True,
     "only a straddle pair at exactly 1000 sees this"),

    ("the activity-description middle boundary drifts 500 -> 400",
     INBER,      "} else if tokens > 500 {", "} else if tokens > 400 {", True, ""),

    # ---- generateNarrative: the token ladder and the joiners ---------------
    ("the narrative 'vast' boundary drifts 2000 -> 2500",
     INBER,      "if stats.TokensUsed > 2000 {", "if stats.TokensUsed > 2500 {", True, ""),

    ("the narrative 'vast' boundary loosens > to >=",
     INBER,      "if stats.TokensUsed > 2000 {", "if stats.TokensUsed >= 2000 {", True, ""),

    ("the narrative 'considerable' boundary drifts 1000 -> 900",
     INBER,      "} else if stats.TokensUsed > 1000 {", "} else if stats.TokensUsed > 900 {", True, ""),

    ("the final list joiner fires for a single-item list too",
     INBER,      "if i == len(activities)-1 && i > 0 {", "if i == len(activities)-1 && i >= 0 {", True, ""),

    # ---- generateJournalTitle ---------------------------------------------
    ("the Great Endeavors threshold drifts 3 -> 4",
     INBER,      "if stats.QuestsCompleted >= 3 {", "if stats.QuestsCompleted >= 4 {", True, ""),

    ("the Great Endeavors threshold narrows >= to >",
     INBER,      "if stats.QuestsCompleted >= 3 {", "if stats.QuestsCompleted > 3 {", True, ""),

    ("the Epic Trials XP threshold drifts 500 -> 600",
     INBER,      "} else if stats.XPGained >= 500 {", "} else if stats.XPGained >= 600 {", True, ""),

    ("the Epic Trials XP threshold narrows >= to >",
     INBER,      "} else if stats.XPGained >= 500 {", "} else if stats.XPGained > 500 {", True, ""),

    # ---- truncateText's own arithmetic ------------------------------------
    ("truncateText's guard narrows <= to <",
     INBER,      "if len(text) <= maxLen {", "if len(text) < maxLen {", True,
     "text at exactly maxLen starts being cut"),

    ("truncateText's ellipsis budget drifts by one byte",
     INBER,      "TruncateAtRuneBoundary(text, maxLen-3)", "TruncateAtRuneBoundary(text, maxLen-4)",
     True, ""),

    # ---- controls ---------------------------------------------------------
    #
    # KNOWN-POSITIVE. Not in this file's scope at all: an achievement threshold
    # that achievements_boundaries_test.go already pins. Both columns must
    # catch it, which is what proves the harness can report CAUGHT and that the
    # PRIOR column is running real tests rather than nothing.
    ("KNOWN-POSITIVE control: an already-pinned achievement threshold drifts",
     HTTPCLIENT, "Level >= 10", "Level >= 11", True,
     "pinned by achievements_boundaries_test.go; BOTH columns must catch it"),

    # KNOWN-NEGATIVE. `threshold := 0` is dead: the switch below it has a
    # default arm, so every path reassigns before the value is read. Neither
    # column may catch this. Without it, a harness that reported CAUGHT for
    # everything would look like a perfect score.
    ("KNOWN-NEGATIVE control: the dead initialiser of threshold drifts 0 -> 99",
     INBER,      "threshold := 0", "threshold := 99", False,
     "a true no-op: the switch's default arm reassigns on every path"),
]


def restore():
    subprocess.run(["git", "checkout", "--"] + TRACKED, cwd=REPO, check=True)
    if MINE_PARKED.exists():
        MINE_PARKED.rename(MINE)


def dirty():
    proc = subprocess.run(["git", "status", "--porcelain"] + TRACKED,
                          cwd=REPO, capture_output=True, text=True)
    return proc.stdout.strip()


def run_suite():
    proc = subprocess.run(["go", "test", "-count=1", PACKAGE],
                          cwd=REPO, capture_output=True, text=True)
    return classify(proc.stdout + proc.stderr, proc.returncode)


if dirty():
    sys.exit("refusing to run: the files under test have uncommitted changes, "
             "and this restores them from git.\n%s" % dirty())
if not MINE.exists():
    sys.exit("refusing to run: %s is missing" % MINE)

# The source file holds a deliberately broken version of itself between the
# write below and the next restore. A try/finally covers SIGINT only, because
# Python turns that one into an exception; SIGTERM and SIGHUP -- what a
# wall-clock cap, systemd and a process-group kill actually send -- would kill
# the process inside that window and leave a mutated tracked file behind,
# looking exactly like ordinary uncommitted work. The handler restores,
# reinstates the disposition it replaced and re-raises, so the process still
# dies BY the signal. SIGKILL cannot be caught, and that gap is named rather
# than papered over.
_previous_handlers = {}


def _restore_and_reraise(signum, frame):
    restore()
    signal.signal(signum, _previous_handlers[signum])
    os.kill(os.getpid(), signum)


for _sig in (signal.SIGINT, signal.SIGTERM, signal.SIGHUP):
    _previous_handlers[_sig] = signal.signal(_sig, _restore_and_reraise)

rows = []
score = 0

try:
    for label, fname, find, replace, expect_mine, note in CASES:
        restore()

        path = REPO / fname
        text = path.read_text()
        occurrences = text.count(find)
        if occurrences != 1:
            print("  SETUP FAIL    %s\n      needle occurs %d times in %s, want exactly 1: %r"
                  % (label, occurrences, fname, find[:70]))
            continue
        path.write_text(text.replace(find, replace, 1))

        # MINE: the package as it now stands.
        mine_verdict, mine_detail = run_suite()

        # PRIOR: the same mutation, with the new file moved aside.
        MINE.rename(MINE_PARKED)
        prior_verdict, _ = run_suite()
        MINE_PARKED.rename(MINE)

        mine_covers = counts_as_coverage(mine_verdict)
        prior_covers = counts_as_coverage(prior_verdict)
        correct = mine_covers == expect_mine
        score += correct

        rows.append((label, prior_verdict, mine_verdict, prior_covers, mine_covers))

        want = "CAUGHT" if expect_mine else "UNNOTICED"
        print("  %s PRIOR=%-14s MINE=%-14s (want MINE %-9s) %s"
              % ("ok  " if correct else "BAD ", prior_verdict, mine_verdict, want, label))
        if mine_detail:
            print("        ↳ %s" % mine_detail)
        if note:
            print("        %s" % note)
        print("        applied %s: %r -> %r"
              % (fname.split("/")[-1], find.strip()[:58], replace.strip()[:58]))
finally:
    restore()
    for _sig, _handler in _previous_handlers.items():
        signal.signal(_sig, _handler)

# ---------------------------------------------------------------------------
# Summary. The number that matters is not "how many did MINE catch" -- it is
# how many MINE catches that PRIOR did not, because a mutation both columns
# catch was already covered and this file added nothing for it.
# ---------------------------------------------------------------------------
closed = [r for r in rows if r[4] and not r[3]]
already = [r for r in rows if r[4] and r[3]]
still_open = [r for r in rows if not r[4] and not r[3]]

print("\n  gaps closed by journal_boundaries_test.go : %d" % len(closed))
print("  already covered before it                : %d" % len(already))
print("  unnoticed by both columns                : %d" % len(still_open))
for label, _, _, _, _ in still_open:
    print("      - %s" % label)

print("\nscore %d/%d" % (score, len(CASES)))
sys.exit(0 if score == len(CASES) else 1)
