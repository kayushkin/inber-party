#!/usr/bin/env python3
"""Score the apostrophe tests by injecting the defects they exist for.

Reading the verdict column (this is the inverted one — see the forty-sixth
nightly pass). Each row injects a real regression and asks which tests notice:

    SOLE DETECTOR  only the named tests failed, and they are the new ones
    REDUNDANT      a pre-existing test also failed; the sibling is named
    UNNOTICED      nothing failed -- the mutation is invisible to the suite
    VOID           the tree did not build, so no test ran at all

VOID is not UNNOTICED. A mutation that orphans an identifier fails to compile,
which is a non-zero exit exactly like a test failure, and scoring that as
"something caught it" is how a scorer reports a comforting lie.
"""

import os
import re
import signal
import subprocess
import sys

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# Messages my tests emit from a t.Fatalf that only guards reachability -- the
# input never got to the code under test. A row that fires ONLY these has not
# scored the mechanism (the forty-second pass).
REACH_GUARDS = [
    "did not return the apostrophe word at all",
    "belongs in the other test",
    "extractKeyTerms returned",
]

HELPER = "internal/textutil/capitalise.go"
QUESTGIVER = "internal/questgiver/questgiver.go"
INBER = "internal/inber/inber.go"

# (label, file, exact_old, exact_new, expectation)
MUTATIONS = [
    ("apostrophe is a separator again (the original defect)",
     HELPER,
     "\t\tif r == '\\'' {",
     "\t\tif r == '\\'' && false {",
     "caught"),

    ("apostrophe never separates (the over-correction)",
     HELPER,
     "\t\t\tatWordStart = !(prevIsWordRune && nextIsWordRune)",
     "\t\t\tatWordStart = !(prevIsWordRune && nextIsWordRune) && false",
     "caught"),

    ("first rune upper-cased instead of title-cased",
     HELPER,
     "b.WriteRune(unicode.ToTitle(r))",
     "b.WriteRune(unicode.ToUpper(r))",
     "caught"),

    ("underscore drifts into being a separator",
     HELPER,
     "\t\tcase r == '_':\n\t\t\treturn true",
     "\t\tcase r == '_':\n\t\t\treturn false",
     "caught"),

    ("invalid UTF-8 replaced with U+FFFD, as strings.Title does",
     HELPER,
     "\t\t\tb.WriteByte(s[i])",
     "\t\t\tb.WriteRune(utf8.RuneError)",
     "caught"),

    ("questgiver extractKeyTerms reverts to strings.Title",
     QUESTGIVER,
     "keyTerms = append(keyTerms, textutil.TitleFirstRuneOfEachWord(word))",
     "keyTerms = append(keyTerms, strings.Title(word))",
     "caught"),

    ("questgiver generic epic name reverts to strings.Title",
     QUESTGIVER,
     "subject = textutil.TitleFirstRuneOfEachWord(strings.ToLower(keyTerms[0]))",
     "subject = strings.Title(strings.ToLower(keyTerms[0]))",
     "caught"),

    # --- controls -------------------------------------------------------
    ("KNOWN POSITIVE: helper returns its input unchanged",
     HELPER,
     "\tvar b strings.Builder",
     "\tif true {\n\t\treturn s\n\t}\n\tvar b strings.Builder",
     "caught"),

    ("KNOWN NEGATIVE: drop the Grow pre-allocation (behaviour-neutral)",
     HELPER,
     "\tb.Grow(len(s))",
     "",
     "unnoticed"),

    ("KNOWN NEGATIVE: inber agent name reverts to strings.Title (unreachable site)",
     INBER,
     'return fmt.Sprintf("The %s Bug Hunt", textutil.TitleFirstRuneOfEachWord(agentName))',
     'return fmt.Sprintf("The %s Bug Hunt", strings.Title(agentName))',
     "unnoticed"),
]

NEW_TESTS = {
    "TestApostrophesDoNotBreakWords",
    "TestEverythingWithoutAnApostropheMatchesStringsTitle",
    "TestAnUnflankedApostropheStillSeparates",
    "TestInvalidUTF8IsCopiedNotReplaced",
    "TestWordStartIsTitleCaseNotUpperCase",
    "TestDifferentialAgainstStringsTitle",
    "TestExtractKeyTermsKeepsAnApostropheInsideAWord",
    "TestGenericEpicNameKeepsAnApostropheInsideAWord",
    "TestWordsWithoutAnApostropheAreUnchanged",
}


def run(cmd):
    return subprocess.run(cmd, cwd=REPO, shell=True, capture_output=True, text=True)


def apply_mutation(path, old, new):
    full = os.path.join(REPO, path)
    src = open(full).read()
    count = src.count(old)
    if count != 1:
        return f"mutation site appears {count} times, not once"
    open(full, "w").write(src.replace(old, new, 1))
    return None


def restore():
    run("git checkout -- .")


def classify():
    """Return (verdict_kind, failing_tests, detail)."""
    build = run("go build ./... && go vet ./...")
    if build.returncode != 0:
        return "VOID", set(), "did not build: " + (build.stderr.strip().splitlines() or [""])[0]

    test = run("go test -count=1 ./...")
    if test.returncode == 0:
        return "UNNOTICED", set(), ""

    out = test.stdout + test.stderr
    if "[build failed]" in out or "cannot use" in out:
        return "VOID", set(), "test binary did not build"

    failing = set(re.findall(r"--- FAIL: (\w+)", out))
    if not failing:
        return "VOID", set(), "non-zero exit with no FAIL line"

    # The forty-second pass: split CAUGHT into assertion-fired and guard-fired.
    # A t.Fatalf that only checks the input reached the code under test is a
    # second sensitive thing in the suite, and a row that trips only those has
    # scored nothing about the mechanism.
    fail_lines = [l for l in out.splitlines() if re.search(r"_test\.go:\d+:", l)]
    guard_only = bool(fail_lines) and all(
        any(g in l for g in REACH_GUARDS) for l in fail_lines
    )

    siblings = failing - NEW_TESTS
    if siblings:
        return "REDUNDANT", failing, "sibling(s): " + ", ".join(sorted(siblings))
    return "SOLE DETECTOR", failing, "guard-fired only" if guard_only else "assertion-fired"


def main():
    restore()
    base = run("go test -count=1 ./...")
    if base.returncode != 0:
        print("BASELINE IS NOT GREEN -- every row below would be meaningless.")
        print((base.stdout + base.stderr)[-2000:])
        return 1
    print("baseline green\n")

    rows = []
    # The mutated file holds a deliberately broken version of itself from the write
    # in the loop below until the restore that follows it, so every way out of that
    # window has to restore — including the ways this script does not choose to
    # take. A killed run left the mutated file behind as ordinary-looking
    # uncommitted work: a semantic edit to a tracked source file, which
    # `git status` reports the same way it reports real work in progress, and which
    # this box's standing rule tells the next agent not to throw away.
    #
    # A try/finally alone does NOT close this, and measuring it is how you find that
    # out. Python raises KeyboardInterrupt for SIGINT, so a finally is on the way
    # out for that one and for nothing else. SIGTERM and SIGHUP kill the process
    # between the write and the restore — and those are exactly what a wall-clock
    # cap, systemd and a process-group kill send. So the one signal a finally covers
    # is the one you press by hand while watching, and the ones it misses are the
    # ones an unattended run actually receives. Measured by kill on this scorer
    # before these handlers existed: SIGTERM left the mutated file behind.
    #
    # The handler restores, reinstates the disposition it replaced and re-raises, so
    # the process dies BY the signal (rc 128+signum). A handler that restores and
    # exits 0 tells every caller a killed run succeeded.
    #
    # SIGKILL cannot be caught by the process that receives it. It is the one gap
    # left here, and it is named rather than papered over.
    previous_handlers = {}

    def restore_and_reraise(signum, frame):
        restore()
        signal.signal(signum, previous_handlers[signum])
        os.kill(os.getpid(), signum)

    for sig in (signal.SIGINT, signal.SIGTERM, signal.SIGHUP):
        previous_handlers[sig] = signal.signal(sig, restore_and_reraise)

    try:
        for label, path, old, new, expectation in MUTATIONS:
            err = apply_mutation(path, old, new)
            if err:
                restore()
                rows.append((label, "VOID", err, expectation))
                continue
            verdict, failing, detail = classify()
            restore()
            rows.append((label, verdict, detail or ", ".join(sorted(failing)), expectation))
    finally:
        restore()
        for sig, handler in previous_handlers.items():
            signal.signal(sig, handler)

    print(f"{'MUTATION':<62} | {'VERDICT':<14} | DETAIL")
    print("-" * 130)
    bad = 0
    for label, verdict, detail, expectation in rows:
        ok = (expectation == "caught" and verdict in ("SOLE DETECTOR", "REDUNDANT")) or (
            expectation == "unnoticed" and verdict == "UNNOTICED")
        if not ok:
            bad += 1
        print(f"{('* ' if not ok else '  ') + label:<62} | {verdict:<14} | {detail[:60]}")

    # The forty-sixth pass: assert the fix is still THERE, not merely that the
    # tree is clean. A restore-after-each-case harness on an uncommitted tree
    # throws the fix away and leaves a clean tree doing it.
    helper = open(os.path.join(REPO, HELPER)).read()
    present = "case r == '_':\n\t\t\treturn true" in helper and "unicode.ToTitle(r)" in helper
    print(f"\nfix still present after the run: {present}")
    status = run("git status --porcelain --untracked-files=no").stdout.strip()
    print(f"tree clean after the run: {not status}")
    print(f"rows disagreeing with expectation: {bad}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
