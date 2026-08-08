#!/usr/bin/env python3
"""Sabotage-score the rune-boundary truncation tests.

A suite that passes tells you nothing on its own. This breaks the fix one
mechanism at a time and records whether the suite goes red. Without that, the
sentence "the tests cover it" is an unmeasured claim.

Rules carried here, each inherited from an unattended pass that lost time to
its absence:

  - The needle must be found, and the file's bytes must actually change. A
    replacement that matched nothing scores UNNOTICED for free, which looks
    exactly like a gap in the tests and is not one.
  - Write a mutation as a DRIFTED COMPARISON, not a deletion, wherever that is
    possible. Deleting the walk-back orphans the utf8 import, and `go test`
    runs vet, so the case reports a compile error instead of a score. A widened
    comparison keeps every identifier live -- and it is the likelier real edit,
    since values drift far more often than checks get deleted.
  - A case that genuinely IS a deletion may carry a second edit to clear the
    orphan it creates.
  - Two controls. A known-positive that every test must catch, and a
    known-NEGATIVE that no test should catch. Without the negative, a harness
    that reports CAUGHT for everything looks perfect while measuring nothing.
  - Restores from git, so the tree must be committed before running.
  - Prints the diff it actually applied, because a row prints the name you gave
    it, not the edit you made.

Run:  python3 scripts/sabotage-truncation.py
"""

import pathlib
import subprocess
import sys

REPO = pathlib.Path(__file__).resolve().parent.parent

TEXTUTIL = "internal/textutil/textutil.go"
INBER = "internal/inber/inber.go"
HTTPCLIENT = "internal/inber/http_client.go"
LOGSTACK = "internal/logstack/logstack.go"
API = "internal/api/api.go"

PACKAGES = ["./internal/textutil/", "./internal/inber/",
            "./internal/logstack/", "./internal/api/"]

# (label, [(file, find, replace)], expect_caught, note)
CASES = [
    ("the walk-back never runs, so the cut splits a rune again",
     [(TEXTUTIL, "for cut > 0 && !utf8.RuneStart(s[cut]) {",
       "for cut > len(s) && !utf8.RuneStart(s[cut]) {")], True, ""),

    ("KNOWN-POSITIVE: the helper trims to nothing instead of to the boundary",
     [(TEXTUTIL, "return s[:cut]", "return s[:0]")], True,
     "every call site's test must notice its text vanishing"),

    ("the walk-back overshoots the boundary by one byte",
     [(TEXTUTIL, "\t}\n\treturn s[:cut]", "\t}\n\treturn s[:cut-1]")], True, ""),

    ("UpperFirstRune skips one byte rather than one rune",
     [(TEXTUTIL, "return string(unicode.ToUpper(first)) + s[width:]",
       "return string(unicode.ToUpper(first)) + s[1:]")], True, ""),

    ("truncateText reverts to a byte cut",
     [(INBER, "return textutil.TruncateAtRuneBoundary(text, maxLen-3) + \"...\"",
       "return text[:maxLen-3] + \"...\"")], True, ""),

    ("the quest-name fallback reverts to a byte cut",
     [(INBER, "text = textutil.TruncateAtRuneBoundary(text, 57) + \"...\"",
       "text = text[:57] + \"...\"")], True, ""),

    ("the HTTP quest description reverts to a byte cut",
     [(HTTPCLIENT, "questDesc = textutil.TruncateAtRuneBoundary(questDesc, 200) + \"...\"",
       "questDesc = questDesc[:200] + \"...\"")], True, ""),

    ("the conversation title reverts to a byte cut",
     [(LOGSTACK, "title = textutil.TruncateAtRuneBoundary(title, 57) + \"...\"",
       "title = title[:57] + \"...\"")], True, ""),

    # A genuine deletion: api.go's only textutil reference is this call, so
    # reverting it orphans the import and vet fails the build. Second edit
    # clears the orphan so the case reports a score instead of a compile error.
    ("the spawn webhook message reverts to a byte cut",
     [(API, "textutil.TruncateAtRuneBoundary(task, 100)", "task[:100]"),
      (API, "\t\"github.com/kayushkin/inber-party/internal/textutil\"\n", "")], True, ""),

    # ---- controls and known gaps -------------------------------------------

    ("KNOWN-NEGATIVE control: the maxBytes guard narrows from <=0 to <0",
     [(TEXTUTIL, "if maxBytes <= 0 {", "if maxBytes < 0 {")], False,
     "a behavioural no-op: maxBytes==0 already yields \"\" through s[:0] below"),

    # Reported, not hidden. These two call sites sit inside functions that read
    # PostgreSQL, so pinning them would mean standing up a database. They are
    # covered by the helper's guarantee, not by a test of their own.
    ("KNOWN GAP: the spawn-message cut in GetConversations reverts",
     [(INBER, "msgContent = textutil.TruncateAtRuneBoundary(msgContent, 100) + \"...\"",
       "msgContent = msgContent[:100] + \"...\"")], False,
     "GetConversations reads PostgreSQL; no test reaches this line"),

    ("KNOWN GAP: the quest description in GetQuests reverts",
     [(INBER, "questDesc = textutil.TruncateAtRuneBoundary(questDesc, 200) + \"...\"",
       "questDesc = questDesc[:200] + \"...\"")], False,
     "GetQuests reads PostgreSQL; the HTTP client's copy of this cut IS pinned"),
]

TRACKED = [TEXTUTIL, INBER, HTTPCLIENT, LOGSTACK, API]


def restore():
    subprocess.run(["git", "checkout", "--"] + TRACKED, cwd=REPO, check=True)


def dirty():
    proc = subprocess.run(["git", "status", "--porcelain"] + TRACKED,
                          cwd=REPO, capture_output=True, text=True)
    return proc.stdout.strip()


if dirty():
    sys.exit("refusing to run: the files under test have uncommitted changes, "
             "and this restores them from git.\n" + dirty())

score = 0
for label, edits, expect_caught, note in CASES:
    restore()

    applied = []
    setup_failed = None
    for fname, find, replace in edits:
        path = REPO / fname
        text = path.read_text()
        if find not in text:
            setup_failed = f"pattern not found in {fname}: {find[:60]!r}"
            break
        new_text = text.replace(find, replace, 1)
        if new_text == text:
            setup_failed = f"edit changed nothing in {fname}"
            break
        path.write_text(new_text)
        applied.append(f"{fname}: {find.strip()[:58]!r} -> {replace.strip()[:58]!r}")

    if setup_failed:
        print(f"  SETUP FAIL    {label}\n      {setup_failed}")
        continue

    proc = subprocess.run(["go", "test", "-count=1"] + PACKAGES,
                          cwd=REPO, capture_output=True, text=True)
    out = proc.stdout + proc.stderr
    if "build failed" in out or "[build failed]" in out or "syntax error" in out:
        verdict = "COMPILE ERROR"
    elif proc.returncode != 0:
        verdict = "CAUGHT"
    else:
        verdict = "UNNOTICED"

    correct = (verdict == "CAUGHT") == expect_caught
    score += correct

    caught_by = ""
    if verdict == "CAUGHT":
        names = sorted({line.split()[2].split("/")[0]
                        for line in out.splitlines() if line.startswith("--- FAIL:")})
        caught_by = " by: " + ",".join(names)

    want = "CAUGHT" if expect_caught else "UNNOTICED"
    print(f"  {'ok  ' if correct else 'BAD '} {verdict:<13} (want {want:<9}) {label}{caught_by}")
    if note:
        print(f"        {note}")
    for line in applied:
        print(f"        applied {line}")

restore()
print(f"\nscore {score}/{len(CASES)}")
sys.exit(0 if score == len(CASES) else 1)
