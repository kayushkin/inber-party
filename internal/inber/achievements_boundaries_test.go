package inber

import (
	"fmt"
	"testing"
)

// Every achievement in computeAchievements unlocks on a numeric comparison, and
// until this file existed not one of those numbers was held in place: moving any
// of them by one left the whole suite green. Each case below sits either exactly
// on a threshold or one step under it, so a literal that drifts in either
// direction reddens a named case.

// hasAchievement reports whether the computed set contains the given achievement id.
func hasAchievement(achievements []RPGAchievement, id string) bool {
	for _, a := range achievements {
		if a.ID == id {
			return true
		}
	}
	return false
}

// questsCompletedByAgent builds count completed quests for one agent. Every quest
// is deliberately short and started mid-morning so that a case aimed at the
// completed-quest thresholds cannot unlock marathon or night_owl by accident.
func questsCompletedByAgent(agentID string, count int) []RPGQuest {
	quests := make([]RPGQuest, 0, count)
	for i := 0; i < count; i++ {
		quests = append(quests, RPGQuest{
			ID:        i + 1,
			AgentID:   agentID,
			Status:    "completed",
			Turns:     1,
			StartedAt: fmt.Sprintf("2026-08-14 10:%02d:00", i%60),
		})
	}
	return quests
}

// questStartedAtHour builds one in-progress quest that began at the given hour.
func questStartedAtHour(agentID string, hour int) []RPGQuest {
	return []RPGQuest{{
		ID:        1,
		AgentID:   agentID,
		Status:    "running",
		Turns:     1,
		StartedAt: fmt.Sprintf("2026-08-14 %02d:30:00", hour),
	}}
}

// questWithTurns builds one quest that ran for the given number of turns.
func questWithTurns(agentID string, turns int) []RPGQuest {
	return []RPGQuest{{
		ID:        1,
		AgentID:   agentID,
		Status:    "completed",
		Turns:     turns,
		StartedAt: "2026-08-14 10:00:00",
	}}
}

func TestComputeAchievementsUnlocksExactlyAtItsThreshold(t *testing.T) {
	const agentID = "claxon"

	cases := []struct {
		name        string
		agent       RPGAgent
		quests      []RPGQuest
		achievement string
		want        bool
	}{
		// Token thresholds. Each pair differs by one token across the line.
		{"999 tokens is not an Apprentice Scribe", RPGAgent{ID: agentID, TotalTokens: 999}, nil, "1k_tokens", false},
		{"1000 tokens is an Apprentice Scribe", RPGAgent{ID: agentID, TotalTokens: 1000}, nil, "1k_tokens", true},
		{"99999 tokens is not a Master Scribe", RPGAgent{ID: agentID, TotalTokens: 99999}, nil, "100k_tokens", false},
		{"100000 tokens is a Master Scribe", RPGAgent{ID: agentID, TotalTokens: 100000}, nil, "100k_tokens", true},
		{"999999 tokens is not an Archmage of Words", RPGAgent{ID: agentID, TotalTokens: 999999}, nil, "1m_tokens", false},
		{"1000000 tokens is an Archmage of Words", RPGAgent{ID: agentID, TotalTokens: 1000000}, nil, "1m_tokens", true},

		// Level thresholds.
		{"level 4 is not a Seasoned Adventurer", RPGAgent{ID: agentID, Level: 4}, nil, "level5", false},
		{"level 5 is a Seasoned Adventurer", RPGAgent{ID: agentID, Level: 5}, nil, "level5", true},
		{"level 9 is not Elite", RPGAgent{ID: agentID, Level: 9}, nil, "level10", false},
		{"level 10 is Elite", RPGAgent{ID: agentID, Level: 10}, nil, "level10", true},

		// Completed-quest thresholds.
		{"9 completed quests is not a Veteran", RPGAgent{ID: agentID}, questsCompletedByAgent(agentID, 9), "10_quests", false},
		{"10 completed quests is a Veteran", RPGAgent{ID: agentID}, questsCompletedByAgent(agentID, 10), "10_quests", true},
		{"49 completed quests is not a Champion", RPGAgent{ID: agentID}, questsCompletedByAgent(agentID, 49), "50_quests", false},
		{"50 completed quests is a Champion", RPGAgent{ID: agentID}, questsCompletedByAgent(agentID, 50), "50_quests", true},

		// Marathon turns the other way round: the comparison is > 30, so 30 is out
		// and 31 is in. A row at 30 is what tells > 30 from >= 30.
		{"30 turns is not a Marathon Runner", RPGAgent{ID: agentID}, questWithTurns(agentID, 30), "marathon", false},
		{"31 turns is a Marathon Runner", RPGAgent{ID: agentID}, questWithTurns(agentID, 31), "marathon", true},

		// The night-owl window is half open on both ends: hour 0 counts, hour 5 does not.
		{"midnight is a Night Owl", RPGAgent{ID: agentID}, questStartedAtHour(agentID, 0), "night_owl", true},
		{"04:30 is a Night Owl", RPGAgent{ID: agentID}, questStartedAtHour(agentID, 4), "night_owl", true},
		{"05:30 is not a Night Owl", RPGAgent{ID: agentID}, questStartedAtHour(agentID, 5), "night_owl", false},

		// First Quest counts this agent's quests, not every quest in the argument.
		// The negative row hands over a quest belonging to somebody else, so it
		// also pins the agent filter that every other case relies on.
		{"no quests at all is not a First Quest", RPGAgent{ID: agentID}, nil, "first_quest", false},
		{"another agent's quest is not a First Quest", RPGAgent{ID: agentID}, questsCompletedByAgent("brigid", 1), "first_quest", false},
		{"one own quest is a First Quest", RPGAgent{ID: agentID}, questsCompletedByAgent(agentID, 1), "first_quest", true},

		// Battle Scarred is a status match rather than a count, and the negative
		// row proves a completed quest cannot stand in for a failed one.
		{"a completed quest is not Battle Scarred", RPGAgent{ID: agentID}, questsCompletedByAgent(agentID, 1), "first_error", false},
		{"a failed quest is Battle Scarred", RPGAgent{ID: agentID},
			[]RPGQuest{{ID: 1, AgentID: agentID, Status: "failed", Turns: 1, StartedAt: "2026-08-14 10:00:00"}},
			"first_error", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			agent := tc.agent
			got := hasAchievement(computeAchievements(&agent, tc.quests), tc.achievement)
			if got != tc.want {
				t.Errorf("computeAchievements: achievement %q present = %v, want %v", tc.achievement, got, tc.want)
			}
		})
	}
}
