#!/usr/bin/env python3
"""Score inber-party's tests for the bounty settlement path by breaking them.

The scoring engine — and the rules it enforces as refusals — lives in
scripts/sabotage.py. This file is only the case list: one edit per mechanism
the suite is meant to pin.

    python3 scripts/sabotage-bounty-settlement.py [--diffs] [--crosstable]

⚠️ **This is the engine's SEVENTH copy, and it is the same blob as the other
six.** md5 `9a81a32e5827b59c1a3093bf88187b17`, taken from git blob
`664f35f475edb9b7d018a28136211bf58a0ff53e` — what
`scheduler/fix/the-scorer-counts-occurrences-not-files`,
`bundle-store/docs/one-sabotage-engine-again`,
`agent-store/test/the-tracked-file-switch-is-unreached`,
`skill-store/test/the-install-path-is-unreached` and
`mailstack/test/the-write-ops-are-unreached` all carry. Diff before editing; an
eighth blob is a fork.

Why this seam is worth a scorer: measured on `main`, a `panic()` on the first
line of **26 of the 27** functions in these two files left `go test ./...`
green. The card named three of the 26.

    internal/db/reputation.go  UpdateReputation      the census row
    internal/db/reputation.go  GetAgentReputation    the only reader of what it writes
    internal/db/reputation.go  InferTaskDomain       routes a bounty to a domain
    internal/db/bounties.go    VerifyBounty          the census row
    internal/db/bounties.go    ResolveDispute        the census row
    internal/db/bounties.go    CreateBounty          derives the tier the gate reads
    internal/db/bounties.go    ClaimBounty           the reputation gate
    internal/db/bounties.go    SubmitWork            the state VerifyBounty acts on
    internal/db/bounties.go    CreateDispute         the state ResolveDispute acts on
    internal/db/bounties.go    WithdrawDispute       the other way a dispute ends
    internal/db/bounties.go    GetBountyByID         read back by three of the above
    internal/db/bounties.go    GetDispute            read back by both resolvers
    internal/db/bounties.go    Scan / Value          required_skills across the driver

That is the **sixth consecutive row** on `e9b0b89c` where the census's function
name under-described the mechanism it belongs to. The 222nd's rule — read the
row's neighbours before scoping the work to its title — has still not once
failed to pay.

⚠️ Read that measurement on `main`, where inber-party's checkout was already
standing. The 223rd's rule is that a reach guard run on whatever branch a
checkout was left on answers about the wrong base, and it fails in the direction
that closes a row without testing it.

What makes this mechanism worth more than a coverage row: **it moves money.**
Every path here ends in `RecordBountyPayout`, which writes a ledger row and
overwrites the agent's gold balance in one transaction. There are two ways to
reach it — an approved verification and a dispute resolved for the claimer — and
each is guarded by exactly one status check standing between it and a second
payout of the same bounty. Neither guard was executed by any test.

⚠️ These tests deliberately do NOT build on db_test.go's `migrateSQLite`. That
helper's `bounties` table has `payout`, `created_by` and `assigned_to`; this code
queries `payout_amount`, `creator_id` and `claimer_id`. Its `reputation` table is
keyed on agent_id alone with `total_score`/`average_rating`; this code keys on
(agent_id, domain) with `score`/`task_count`/`success_rate`. The two are
different schemas, and a test written against the helper would pin a table
nothing under test reads. The suite translates the production migration in
db.go instead.
"""

import sys

sys.path.insert(0, str(__import__("pathlib").Path(__file__).resolve().parent))

from sabotage import REPO, Case, score  # noqa: E402

TARGETS = [REPO / "internal" / "db" / "bounties.go", REPO / "internal" / "db" / "reputation.go"]
PACKAGES = ["./internal/db"]

CASES = [
    # ---- UpdateReputation: the arithmetic ----
    Case(
        "a success is worth a flat bonus rather than one scaled by the success rate",
        [("\t\tscoreChange = 5 + int(10*newSuccessRate)",
          "\t\tscoreChange = 5 + int(5*newSuccessRate)")],
    ),
    Case(
        "a failure costs twice what it should",
        [("\t\tscoreChange = -10 - int(5*(1.0-newSuccessRate))",
          "\t\tscoreChange = -10 - int(10*(1.0-newSuccessRate))")],
    ),
    Case(
        "the score floor is dropped, so a failing agent goes negative",
        [("\tif newScore < 0 {", "\tif newScore < -10000 {")],
    ),
    Case(
        "the score ceiling is dropped, so reputation grows without bound",
        [("\tif newScore > 1000 {", "\tif newScore > 100000 {")],
    ),
    Case(
        "the task counter advances by two, so every rate is computed against the wrong total",
        [("\tnewTaskCount := taskCount + 1", "\tnewTaskCount := taskCount + 2")],
    ),
    Case(
        "a success contributes twice its weight to the running average",
        [("\t\tnewSuccessRate = (successRate*float64(taskCount) + 1.0) / float64(newTaskCount)",
          "\t\tnewSuccessRate = (successRate*float64(taskCount) + 2.0) / float64(newTaskCount)")],
    ),
    Case(
        "a failure keeps the old rate instead of diluting it",
        [("\t\tnewSuccessRate = (successRate * float64(taskCount)) / float64(newTaskCount)",
          "\t\tnewSuccessRate = successRate")],
    ),
    Case(
        "a new agent is seeded at a different starting score",
        [("\t\tVALUES ($1, $2, 100, 0, 1.0, CURRENT_TIMESTAMP)",
          "\t\tVALUES ($1, $2, 200, 0, 1.0, CURRENT_TIMESTAMP)")],
    ),

    # ---- GetAgentReputation ----
    Case(
        "an agent's domains come back worst-first",
        [("\t\tORDER BY score DESC\n\t`, agentID)", "\t\tORDER BY score ASC\n\t`, agentID)")],
    ),

    # ---- InferTaskDomain: the router the claim gate reads ----
    Case(
        "an unrecognised task falls back to a real domain instead of general",
        [('\tbestDomain := "general"', '\tbestDomain := "coding"')],
    ),
    Case(
        "the description is ignored, so only the title routes a bounty",
        [('\ttext := strings.ToLower(taskName + " " + taskDescription)',
          "\ttext := strings.ToLower(taskName)")],
    ),
    Case(
        "the router stops folding case, so a title in caps routes to general",
        [('\ttext := strings.ToLower(taskName + " " + taskDescription)',
          '\ttext := taskName + " " + taskDescription')],
    ),

    # ---- CreateBounty: the tier the claim gate is keyed on ----
    Case(
        "the legendary boundary excludes the payout that defines it",
        [("\tif b.PayoutAmount >= 500 {", "\tif b.PayoutAmount > 500 {")],
    ),
    Case(
        "the gold boundary excludes the payout that defines it",
        [("\t} else if b.PayoutAmount >= 200 {", "\t} else if b.PayoutAmount > 200 {")],
    ),
    Case(
        "the silver boundary excludes the payout that defines it",
        [("\t} else if b.PayoutAmount >= 50 {", "\t} else if b.PayoutAmount > 50 {")],
    ),
    Case(
        "an unpaid bounty defaults to a tier above bronze",
        [('\ttier := "bronze"', '\ttier := "silver"')],
    ),
    Case(
        "the derived tier is returned to the caller but never stored",
        [("\t\tb.PayoutAmount, b.CreatorID, requiredSkills, tier, b.Deadline).Scan(",
          '\t\tb.PayoutAmount, b.CreatorID, requiredSkills, "bronze", b.Deadline).Scan(')],
    ),

    # ---- ClaimBounty: the reputation gate ----
    Case(
        "the legendary tier stops demanding a high reputation",
        [("\t\tminReputationRequired = 750", "\t\tminReputationRequired = 50")],
    ),
    Case(
        "the silver tier demands an unreachable reputation",
        [("\t\tminReputationRequired = 250", "\t\tminReputationRequired = 2500")],
    ),
    Case(
        "the gate refuses on EITHER condition rather than both, so experience stops counting",
        [("\tif agentReputation.Score < minReputationRequired && agentReputation.TaskCount < 3 {",
          "\tif agentReputation.Score < minReputationRequired || agentReputation.TaskCount < 3 {")],
    ),
    Case(
        "the experience door needs far more tasks than it should",
        [("\tif agentReputation.Score < minReputationRequired && agentReputation.TaskCount < 3 {",
          "\tif agentReputation.Score < minReputationRequired && agentReputation.TaskCount < 30 {")],
    ),
    Case(
        "an agent with no reputation is treated as a top-scoring one",
        [("\t\tagentReputation = &Reputation{Score: 100, TaskCount: 0, SuccessRate: 1.0}",
          "\t\tagentReputation = &Reputation{Score: 1000, TaskCount: 0, SuccessRate: 1.0}")],
    ),
    # Dominated by construction, and declared rather than chased. ClaimBounty
    # checks `status = 'open'` TWICE — once in the SELECT that opens the
    # function, once in the UPDATE that closes it — and the SELECT refuses a
    # claimed bounty before the UPDATE is ever reached. The UPDATE's copy is
    # not redundant: it is the compare-and-swap that closes the check-then-act
    # race between two agents claiming at once. No sequential test can separate
    # them, which is the 221st's shape — check whether the mutation is
    # observable at all before reading an UNNOTICED row as a coverage hole.
    Case(
        "the claim writes over whatever the bounty's current state is, so a claimed bounty is re-claimable",
        [("\t\tWHERE id = $2 AND status = 'open'\n\t`\n\t\n\tresult, err := db.Exec(query, claimerID, bountyID)",
          "\t\tWHERE id = $2\n\t`\n\t\n\tresult, err := db.Exec(query, claimerID, bountyID)")],
        expected_unnoticed="the SELECT at the top of ClaimBounty rejects a non-open bounty before this UPDATE runs, so sequentially the two guards are the same guard; this copy only closes the concurrent check-then-act race",
    ),

    # ---- SubmitWork ----
    Case(
        "work can be submitted against a bounty nobody claimed",
        [("\t\tWHERE id = $2 AND status = 'claimed'", "\t\tWHERE id = $2")],
    ),

    # ---- VerifyBounty ----
    Case(
        "approval and rejection are swapped",
        [('\tstatus := "rejected"', '\tstatus := "__swapped_completed"'),
         ('\t\tstatus = "completed"', '\t\tstatus = "rejected"')],
    ),
    Case(
        "verification acts on a bounty in any state, so an unsubmitted bounty can be approved",
        [("\t\tWHERE id = $4 AND status = 'submitted'", "\t\tWHERE id = $4")],
    ),
    # The second edit is the engine's orphan escape: dropping the only use of
    # `now` makes this a compile error rather than a score, which is the 223rd's
    # warning and the 225th's generalisation of it.
    Case(
        "an approved bounty is never stamped complete",
        [("\t\tcompletedAt = &now", "\t\tcompletedAt = nil"),
         ('\tstatus := "rejected"\n\tcompletedAt := (*time.Time)(nil)\n\tnow := time.Now()',
          '\tstatus := "rejected"\n\tcompletedAt := (*time.Time)(nil)\n\tnow := time.Now()\n\t_ = now')],
    ),
    Case(
        "the payout is recorded against the bounty's creator rather than its claimer",
        [("\t\t\terr = db.RecordBountyPayout(bountyID, *bounty.ClaimerID, bounty.PayoutAmount)",
          "\t\t\terr = db.RecordBountyPayout(bountyID, bounty.CreatorID, bounty.PayoutAmount)")],
    ),
    Case(
        "an approved bounty pays nothing",
        [("\t\t\terr = db.RecordBountyPayout(bountyID, *bounty.ClaimerID, bounty.PayoutAmount)",
          "\t\t\terr = db.RecordBountyPayout(bountyID, *bounty.ClaimerID, 0)")],
    ),
    Case(
        "a paid bounty is left reading as merely completed",
        [("\t\t\t\tUPDATE bounties \n\t\t\t\tSET status = 'paid', updated_at = CURRENT_TIMESTAMP \n\t\t\t\tWHERE id = $1\n\t\t\t`, bountyID)\n\t\t\tif err != nil {\n\t\t\t\treturn fmt.Errorf(\"failed to mark bounty as paid: %w\", err)\n\t\t\t}\n\t\t}\n\t}\n\t\n\treturn nil\n}\n\n// payoutBounty",
          "\t\t\t\tUPDATE bounties \n\t\t\t\tSET status = 'completed', updated_at = CURRENT_TIMESTAMP \n\t\t\t\tWHERE id = $1\n\t\t\t`, bountyID)\n\t\t\tif err != nil {\n\t\t\t\treturn fmt.Errorf(\"failed to mark bounty as paid: %w\", err)\n\t\t\t}\n\t\t}\n\t}\n\t\n\treturn nil\n}\n\n// payoutBounty")],
    ),

    # ---- CreateDispute ----
    Case(
        "a dispute can be filed against a bounty that was never rejected",
        [("\t\tWHERE id = $1 AND status = 'rejected'", "\t\tWHERE id = $1")],
    ),
    Case(
        "anyone may dispute a bounty, not only the agent who worked it",
        [("\tif bounty.ClaimerID == nil || *bounty.ClaimerID != d.ClaimerID {",
          "\tif bounty.ClaimerID == nil {")],
    ),
    Case(
        "the dispute's creator is taken from the caller rather than from the bounty",
        [("\td.CreatorID = bounty.CreatorID", "\td.CreatorID = d.ClaimerID")],
    ),
    Case(
        "filing a dispute leaves the bounty reading as plainly rejected",
        [("\t\tSET status = 'disputed', updated_at = CURRENT_TIMESTAMP \n\t\tWHERE id = $1\n\t`, d.BountyID)",
          "\t\tSET status = 'rejected', updated_at = CURRENT_TIMESTAMP \n\t\tWHERE id = $1\n\t`, d.BountyID)")],
    ),

    # ---- ResolveDispute ----
    Case(
        "an already-resolved dispute can be resolved again, and paid again",
        [('\tif dispute.Status != "open" {', '\tif dispute.Status == "__never" {')],
    ),
    Case(
        "the two resolution outcomes are swapped",
        [('\tnewStatus := "resolved_against"', '\tnewStatus := "__swapped"'),
         ('\t\tnewStatus = "resolved_in_favor"', '\t\tnewStatus = "resolved_against"')],
    ),
    Case(
        "a win for the claimer is recorded in the notes as a loss",
        [('"\\n\\n[DISPUTE RESOLVED IN FAVOR OF CLAIMER] "+resolution',
          '"\\n\\n[DISPUTE RESOLVED AGAINST CLAIMER] "+resolution')],
    ),
    Case(
        "the resolution replaces the rejection notes instead of appending to them",
        [("\t\t\t    verification_notes = COALESCE(verification_notes, '') || $2",
          "\t\t\t    verification_notes = $2")],
    ),
    Case(
        "losing a dispute reopens the bounty instead of leaving it rejected",
        [("\t\t\tSET status = 'rejected', updated_at = CURRENT_TIMESTAMP,\n\t\t\t    verification_notes = COALESCE(verification_notes, '') || $1",
          "\t\t\tSET status = 'open', updated_at = CURRENT_TIMESTAMP,\n\t\t\t    verification_notes = COALESCE(verification_notes, '') || $1")],
    ),
    Case(
        "winning a dispute pays nothing",
        [("\t\t\terr = db.RecordBountyPayout(dispute.BountyID, *bounty.ClaimerID, bounty.PayoutAmount)",
          "\t\t\terr = db.RecordBountyPayout(dispute.BountyID, *bounty.ClaimerID, 0)")],
    ),
    Case(
        "a missing dispute is reported with a message that does not say what is missing",
        [('\t\treturn fmt.Errorf("dispute not found")', '\t\treturn fmt.Errorf("not found")')],
    ),

    # ---- WithdrawDispute ----
    Case(
        "the withdrawal is gated on the bounty's creator rather than its claimer",
        [("\t\tWHERE id = $1 AND claimer_id = $2 AND status = 'open'",
          "\t\tWHERE id = $1 AND creator_id = $2 AND status = 'open'")],
    ),
    Case(
        "a resolved dispute can still be withdrawn, dragging a paid bounty back to rejected",
        [("\t\tWHERE id = $1 AND claimer_id = $2 AND status = 'open'",
          "\t\tWHERE id = $1 AND claimer_id = $2")],
    ),
    Case(
        "withdrawing a dispute reopens the bounty instead of restoring the rejection",
        [("\t\tSET status = 'rejected', updated_at = CURRENT_TIMESTAMP \n\t\tWHERE id = $1\n\t`, dispute.BountyID)",
          "\t\tSET status = 'open', updated_at = CURRENT_TIMESTAMP \n\t\tWHERE id = $1\n\t`, dispute.BountyID)")],
    ),

    # ---- StringSlice and the two not-found readers ----
    Case(
        "a value the skills column cannot hold is swallowed instead of refused",
        [("\t\treturn fmt.Errorf(\"cannot scan %T into StringSlice\", value)", "\t\treturn nil")],
    ),
    Case(
        "a missing bounty is reported as an error rather than as no bounty",
        [("\t\tif err == sql.ErrNoRows {\n\t\t\treturn nil, nil\n\t\t}\n\t\treturn nil, fmt.Errorf(\"failed to get bounty: %w\", err)",
          "\t\treturn nil, fmt.Errorf(\"failed to get bounty: %w\", err)")],
    ),
    Case(
        "a missing dispute is reported as an error rather than as no dispute",
        [("\t\tif err == sql.ErrNoRows {\n\t\t\treturn nil, nil\n\t\t}\n\t\treturn nil, fmt.Errorf(\"failed to get dispute: %w\", err)",
          "\t\treturn nil, fmt.Errorf(\"failed to get dispute: %w\", err)")],
    ),

    # ---- controls ----
    # Known-positive: the column every bounty in this suite is written through.
    # Drifts the name rather than deleting the statement, per the case rules —
    # every test here creates a bounty, so it has to redden.
    Case(
        "CONTROL known-positive: bounties are written to a neighbouring column name",
        [("\t\tINSERT INTO bounties (title, description, requirements, payout_amount, ",
          "\t\tINSERT INTO bounties (title, description, requirements, payout_amount_v2, ")],
    ),
    # Known-negative: the WORDING of the already-claimed refusal, taken at the
    # SELECT that actually raises it. ⚠️ The obvious site — the identical
    # sentence under `rowsAffected == 0` — would be a WORTHLESS negative: that
    # branch is unreachable sequentially for exactly the reason declared on the
    # re-claim case above, so it would score UNNOTICED whether or not the suite
    # was any good. This site IS reached: TestClaimBountyRefusesAnAlreadyClaimed
    # Bounty runs straight through it, and the suite asserts that the claim was
    # refused and that the first claimer still holds the bounty — never the
    # sentence the refusal is phrased in.
    Case(
        "CONTROL known-negative: the already-claimed refusal is reworded",
        [('\tif err != nil {\n\t\treturn fmt.Errorf("bounty not found or already claimed")\n\t}',
          '\tif err != nil {\n\t\treturn fmt.Errorf("this bounty has already been taken by another agent")\n\t}')],
        expected_unnoticed="the suite asserts the claim was refused and that the first claimer still holds the bounty, never the sentence the refusal is phrased in",
    ),
]


if __name__ == "__main__":
    sys.exit(score(TARGETS, PACKAGES, CASES))
