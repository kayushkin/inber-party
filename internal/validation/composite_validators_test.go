package validation

import (
	"sort"
	"strings"
	"testing"
)

// The three composite validators pinned here — ValidateRatingData, ValidateTaskData and
// ValidatePartyData — had no test of any kind. Deleting any single guard line inside them
// left the whole suite green, which is the state card 08f0451b's last paragraph describes
// for the rating range: the guard exists, and nothing reads it.
//
// These tests assert the FIELD and the CODE of every error a call produces, not the count.
// A count-only assertion (the style of TestValidateBountyData and TestValidateAgentData
// above) cannot tell a guard that fired from a different guard firing on another field, so
// it passes for the wrong reason whenever a limit moves between two fields.

// fieldCode identifies which guard fired, which is what a count cannot say.
type fieldCode struct {
	Field string
	Code  string
}

func (fc fieldCode) String() string { return fc.Field + "/" + fc.Code }

// assertValidationErrors compares the full set of errors a validator produced against the
// full set expected, so an unexpected error fails just as loudly as a missing one.
func assertValidationErrors(t *testing.T, v *Validator, want ...fieldCode) {
	t.Helper()

	var got []fieldCode
	for _, err := range v.GetErrors() {
		got = append(got, fieldCode{Field: err.Field, Code: err.Code})
	}

	render := func(fcs []fieldCode) string {
		if len(fcs) == 0 {
			return "(none)"
		}
		parts := make([]string, 0, len(fcs))
		for _, fc := range fcs {
			parts = append(parts, fc.String())
		}
		sort.Strings(parts)
		return strings.Join(parts, ", ")
	}

	if gotStr, wantStr := render(got), render(want); gotStr != wantStr {
		t.Errorf("validation errors: got [%s], want [%s]", gotStr, wantStr)
		for _, err := range v.GetErrors() {
			t.Logf("  error: %s", err.Error())
		}
	}
}

func TestValidateRatingData(t *testing.T) {
	const validComment = "solid work"

	tests := []struct {
		name     string
		rating   int
		comment  string
		bountyID int
		raterID  int
		ratedID  int
		want     []fieldCode
	}{
		{name: "a valid five-star rating passes", rating: 5, comment: validComment, bountyID: 1, raterID: 2, ratedID: 3},

		// The range guard, straddled at both ends. 1 and 5 must pass and 0 and 6 must fail,
		// so widening the range in either direction reddens a case rather than going unnoticed.
		{name: "the lowest legal rating passes", rating: 1, comment: validComment, bountyID: 1, raterID: 2, ratedID: 3},
		{name: "the highest legal rating passes", rating: 5, comment: validComment, bountyID: 1, raterID: 2, ratedID: 3},
		{name: "a rating below the range is refused", rating: 0, comment: validComment, bountyID: 1, raterID: 2, ratedID: 3,
			want: []fieldCode{{"rating", "range"}}},
		{name: "a rating above the range is refused", rating: 6, comment: validComment, bountyID: 1, raterID: 2, ratedID: 3,
			want: []fieldCode{{"rating", "range"}}},
		{name: "a negative rating is refused", rating: -1, comment: validComment, bountyID: 1, raterID: 2, ratedID: 3,
			want: []fieldCode{{"rating", "range"}}},
		{name: "a rating far above the range is refused", rating: 9, comment: validComment, bountyID: 1, raterID: 2, ratedID: 3,
			want: []fieldCode{{"rating", "range"}}},

		// The comment length guard, straddled on the 1000-character limit.
		{name: "a comment at the length limit passes", rating: 4, comment: strings.Repeat("a", 1000), bountyID: 1, raterID: 2, ratedID: 3},
		{name: "a comment past the length limit is refused", rating: 4, comment: strings.Repeat("a", 1001), bountyID: 1, raterID: 2, ratedID: 3,
			want: []fieldCode{{"comment", "max_length"}}},
		{name: "an empty comment passes", rating: 4, comment: "", bountyID: 1, raterID: 2, ratedID: 3},

		// Each id is guarded separately, so each gets its own case. A single case with all
		// three missing could not tell one guard's absence from another's.
		{name: "a missing bounty id is refused", rating: 4, comment: validComment, bountyID: 0, raterID: 2, ratedID: 3,
			want: []fieldCode{{"bounty_id", "required"}}},
		{name: "a missing rater id is refused", rating: 4, comment: validComment, bountyID: 1, raterID: 0, ratedID: 3,
			want: []fieldCode{{"rater_id", "required"}}},
		{name: "a missing rated id is refused", rating: 4, comment: validComment, bountyID: 1, raterID: 2, ratedID: 0,
			want: []fieldCode{{"rated_id", "required"}}},

		{name: "every guard reports independently", rating: 42, comment: strings.Repeat("a", 1001), bountyID: 0, raterID: 0, ratedID: 0,
			want: []fieldCode{
				{"rating", "range"}, {"comment", "max_length"},
				{"bounty_id", "required"}, {"rater_id", "required"}, {"rated_id", "required"},
			}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateRatingData(tt.rating, tt.comment, tt.bountyID, tt.raterID, tt.ratedID)
			assertValidationErrors(t, v, tt.want...)
		})
	}
}

func TestValidateTaskData(t *testing.T) {
	const (
		validName        = "Clear the goblin camp"
		validDescription = "A test description"
	)

	tests := []struct {
		name        string
		taskName    string
		description string
		difficulty  string
		xpReward    int
		goldReward  int
		want        []fieldCode
	}{
		{name: "valid task data passes", taskName: validName, description: validDescription, difficulty: "medium", xpReward: 100, goldReward: 50},

		{name: "an empty name is refused", taskName: "", description: validDescription, difficulty: "medium", xpReward: 100, goldReward: 50,
			want: []fieldCode{{"name", "required"}}},
		{name: "a name at the length limit passes", taskName: strings.Repeat("a", 200), description: validDescription, difficulty: "medium", xpReward: 100, goldReward: 50},
		{name: "a name past the length limit is refused", taskName: strings.Repeat("a", 201), description: validDescription, difficulty: "medium", xpReward: 100, goldReward: 50,
			want: []fieldCode{{"name", "max_length"}}},

		{name: "an empty description is refused", taskName: validName, description: "", difficulty: "medium", xpReward: 100, goldReward: 50,
			want: []fieldCode{{"description", "required"}}},
		{name: "a description past the length limit is refused", taskName: validName, description: strings.Repeat("a", 2001), difficulty: "medium", xpReward: 100, goldReward: 50,
			want: []fieldCode{{"description", "max_length"}}},

		// difficulty is a required enum, so an empty value and an unlisted value fail with
		// different codes. Asserting the code is what separates the two.
		{name: "a missing difficulty is refused as required", taskName: validName, description: validDescription, difficulty: "", xpReward: 100, goldReward: 50,
			want: []fieldCode{{"difficulty", "required"}}},
		{name: "an unlisted difficulty is refused as an invalid option", taskName: validName, description: validDescription, difficulty: "impossible", xpReward: 100, goldReward: 50,
			want: []fieldCode{{"difficulty", "invalid_option"}}},
		{name: "a numeric difficulty is listed and passes", taskName: validName, description: validDescription, difficulty: "3", xpReward: 100, goldReward: 50},

		// Both reward ranges straddled at both ends, and they carry different limits — so a
		// limit copied from one field to the other reddens instead of passing quietly.
		{name: "the lowest legal xp reward passes", taskName: validName, description: validDescription, difficulty: "medium", xpReward: 1, goldReward: 50},
		{name: "the highest legal xp reward passes", taskName: validName, description: validDescription, difficulty: "medium", xpReward: 10000, goldReward: 50},
		{name: "an xp reward below the range is refused", taskName: validName, description: validDescription, difficulty: "medium", xpReward: 0, goldReward: 50,
			want: []fieldCode{{"xp_reward", "range"}}},
		{name: "an xp reward above the range is refused", taskName: validName, description: validDescription, difficulty: "medium", xpReward: 10001, goldReward: 50,
			want: []fieldCode{{"xp_reward", "range"}}},
		{name: "the lowest legal gold reward passes", taskName: validName, description: validDescription, difficulty: "medium", xpReward: 100, goldReward: 1},
		{name: "the highest legal gold reward passes", taskName: validName, description: validDescription, difficulty: "medium", xpReward: 100, goldReward: 50000},
		{name: "a gold reward below the range is refused", taskName: validName, description: validDescription, difficulty: "medium", xpReward: 100, goldReward: 0,
			want: []fieldCode{{"gold_reward", "range"}}},
		{name: "a gold reward above the range is refused", taskName: validName, description: validDescription, difficulty: "medium", xpReward: 100, goldReward: 50001,
			want: []fieldCode{{"gold_reward", "range"}}},

		{name: "every guard reports independently", taskName: "", description: "", difficulty: "impossible", xpReward: 0, goldReward: 0,
			want: []fieldCode{
				{"name", "required"}, {"description", "required"}, {"difficulty", "invalid_option"},
				{"xp_reward", "range"}, {"gold_reward", "range"},
			}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateTaskData(tt.taskName, tt.description, tt.difficulty, tt.xpReward, tt.goldReward)
			assertValidationErrors(t, v, tt.want...)
		})
	}
}

func TestValidatePartyData(t *testing.T) {
	const validPartyName = "The Iron Band"

	tests := []struct {
		name        string
		partyName   string
		description string
		leaderID    int
		want        []fieldCode
	}{
		{name: "valid party data passes", partyName: validPartyName, description: "A test description", leaderID: 1},

		{name: "an empty name is refused", partyName: "", description: "A test description", leaderID: 1,
			want: []fieldCode{{"name", "required"}}},
		{name: "a name at the length limit passes", partyName: strings.Repeat("a", 100), description: "A test description", leaderID: 1},
		{name: "a name past the length limit is refused", partyName: strings.Repeat("a", 101), description: "A test description", leaderID: 1,
			want: []fieldCode{{"name", "max_length"}}},

		// The description is optional but bounded, and its limit differs from the name's.
		{name: "an empty description passes", partyName: validPartyName, description: "", leaderID: 1},
		{name: "a description at the length limit passes", partyName: validPartyName, description: strings.Repeat("a", 500), leaderID: 1},
		{name: "a description past the length limit is refused", partyName: validPartyName, description: strings.Repeat("a", 501), leaderID: 1,
			want: []fieldCode{{"description", "max_length"}}},

		{name: "a missing leader id is refused", partyName: validPartyName, description: "A test description", leaderID: 0,
			want: []fieldCode{{"leader_id", "required"}}},
		{name: "a negative leader id is refused", partyName: validPartyName, description: "A test description", leaderID: -1,
			want: []fieldCode{{"leader_id", "required"}}},

		{name: "every guard reports independently", partyName: "", description: strings.Repeat("a", 501), leaderID: 0,
			want: []fieldCode{
				{"name", "required"}, {"description", "max_length"}, {"leader_id", "required"},
			}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidatePartyData(tt.partyName, tt.description, tt.leaderID)
			assertValidationErrors(t, v, tt.want...)
		})
	}
}
