package inber

import (
	"fmt"
	"testing"
)

func TestLevelForXPComprehensive(t *testing.T) {
	tests := []struct {
		name        string
		xp          int
		wantLevel   int
		wantXPToNext int
	}{
		// Level 1 tests (0-99 XP)
		{"Level 1 start", 0, 1, 100},
		{"Level 1 middle", 50, 1, 50},
		{"Level 1 end", 99, 1, 1},
		
		// Level 2 tests (100-299 XP total)
		{"Level 2 start", 100, 2, 200},
		{"Level 2 middle", 200, 2, 100},
		{"Level 2 end", 299, 2, 1},
		
		// Level 3 tests (300-599 XP total)
		{"Level 3 start", 300, 3, 300},
		{"Level 3 middle", 450, 3, 150},
		{"Level 3 end", 599, 3, 1},
		
		// Level 4 tests (600-999 XP total)
		{"Level 4 start", 600, 4, 400},
		{"Level 4 middle", 800, 4, 200},
		{"Level 4 end", 999, 4, 1},
		
		// Level 5 tests (1000-1499 XP total)
		{"Level 5 start", 1000, 5, 500},
		{"Level 5 end", 1499, 5, 1},
		
		// Higher level tests
		{"Level 10", 4500, 10, 1000}, // Level 10 needs 4500 total, next level needs 1000 more
		{"Level 15", 10500, 15, 1500}, // Level 15 needs 10500 total, next level needs 1500 more
		
		// Edge case: very high XP should cap at level 99
		{"Level cap test", 1000000, 99, 0},
		
		// Verify the progression formula: level N requires sum(1 to N-1) * 100 XP
		{"Formula verification level 6", 1500, 6, 600}, // sum(1 to 5) * 100 = 15 * 100 = 1500
		{"Formula verification level 7", 2100, 7, 700},   // sum(1 to 6) * 100 = 21 * 100 = 2100
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLevel, gotXPToNext := levelForXP(tt.xp)
			if gotLevel != tt.wantLevel {
				t.Errorf("levelForXP(%d) level = %d, want %d", tt.xp, gotLevel, tt.wantLevel)
			}
			if gotXPToNext != tt.wantXPToNext {
				t.Errorf("levelForXP(%d) xpToNext = %d, want %d", tt.xp, gotXPToNext, tt.wantXPToNext)
			}
		})
	}
}

func TestLevelForXPProgression(t *testing.T) {
	// Test that each level requires progressively more XP
	prevTotalXP := 0
	for level := 1; level <= 10; level++ {
		totalXPForLevel := 0
		// Calculate total XP needed for this level: sum(1 to level-1) * 100
		for i := 1; i < level; i++ {
			totalXPForLevel += i * 100
		}
		
		gotLevel, xpToNext := levelForXP(totalXPForLevel)
		if gotLevel != level {
			t.Errorf("Expected XP %d to be level %d, got level %d", totalXPForLevel, level, gotLevel)
		}
		
		// At exact level threshold, xpToNext should equal level * 100
		expectedXPToNext := level * 100
		if xpToNext != expectedXPToNext {
			t.Errorf("At level %d start (XP %d), expected xpToNext %d, got %d", 
				level, totalXPForLevel, expectedXPToNext, xpToNext)
		}
		
		// Verify progression (each level should require more XP than previous)
		if level > 1 && totalXPForLevel <= prevTotalXP {
			t.Errorf("Level %d requires %d XP, which is not more than level %d requiring %d XP", 
				level, totalXPForLevel, level-1, prevTotalXP)
		}
		
		prevTotalXP = totalXPForLevel
	}
}

func TestClassForComprehensive(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		wantClass string
		wantEmoji string
		wantTitle string
	}{
		// Inber agents
		{"Claxon overlord", "claxon", "Overlord", "♚", "the All-Seeing"},
		{"Claxon case insensitive", "CLAXON", "Overlord", "♚", "the All-Seeing"},
		{"Fionn engineer", "fionn", "Engineer", "⚙️", "the Builder"},
		{"Brigid artificer", "brigid", "Artificer", "🔨", "the Radiant"},
		{"Oisin bard", "oisin", "Bard", "🎵", "the Storyteller"},
		{"Manannan ranger", "manannan", "Ranger", "🌊", "the Voyager"},
		{"Ogma scribe", "ogma", "Scribe", "📜", "the Chronicler"},
		{"Scathach shadow", "scathach", "Shadow", "🗡️", "the Unseen"},
		{"Goibniu smith", "goibniu", "Smith", "⚒️", "the Forgemaster"},
		{"Bench gladiator", "bench", "Gladiator", "🏛️", "the Proven"},
		{"Bran scout", "bran", "Scout", "🐕", "the Loyal"},
		
		// OpenClaw agents
		{"Main sovereign", "main", "Sovereign", "👑", "the Wise"},
		{"Kayushkin artificer", "kayushkin", "Artificer", "🔨", "the Maker"},
		{"Si bard", "si", "Bard", "🎵", "the Melodic"},
		{"Downloadstack ranger", "downloadstack", "Ranger", "🌊", "the Fetcher"},
		{"Claxon-android shadow", "claxon-android", "Shadow", "📱", "the Mobile"},
		{"Inber engineer", "inber", "Engineer", "⚙️", "the Orchestrator"},
		{"Forge smith", "forge", "Smith", "⚒️", "the Forgemaster"},
		{"Logstack scribe", "logstack", "Scribe", "📜", "the Recorder"},
		{"Healthcheck cleric", "healthcheck", "Cleric", "🏥", "the Healer"},
		{"Inber-party jester", "inber-party", "Jester", "🃏", "the Entertainer"},
		{"Agent-bench gladiator", "agent-bench", "Gladiator", "🏛️", "the Tested"},
		{"Argraphments sage", "argraphments", "Sage", "📚", "the Learned"},
		{"Keyboard tinker", "keyboard", "Tinker", "🎹", "the Melodic"},
		{"Claxon-watch sentinel", "claxon-watch", "Sentinel", "⌚", "the Watchful"},
		{"Run ranger", "run", "Ranger", "🏹", "the Swift"},
		
		// Unknown agents (defaults)
		{"Unknown agent", "unknown-agent", "Adventurer", "⚔️", "the Unknown"},
		{"Empty string", "", "Adventurer", "⚔️", "the Unknown"},
		{"Random name", "random123", "Adventurer", "⚔️", "the Unknown"},
		
		// Case sensitivity tests
		{"Mixed case", "ClAxOn", "Overlord", "♚", "the All-Seeing"},
		{"Upper case", "FIONN", "Engineer", "⚙️", "the Builder"},
		{"Lower case", "brigid", "Artificer", "🔨", "the Radiant"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClass, gotEmoji, gotTitle := classFor(tt.agentName)
			if gotClass != tt.wantClass {
				t.Errorf("classFor(%q) class = %q, want %q", tt.agentName, gotClass, tt.wantClass)
			}
			if gotEmoji != tt.wantEmoji {
				t.Errorf("classFor(%q) emoji = %q, want %q", tt.agentName, gotEmoji, tt.wantEmoji)
			}
			if gotTitle != tt.wantTitle {
				t.Errorf("classFor(%q) title = %q, want %q", tt.agentName, gotTitle, tt.wantTitle)
			}
		})
	}
}

func TestAgentClassesMapping(t *testing.T) {
	// Test that all entries in agentClasses map are valid and complete
	for agentName, classInfo := range agentClasses {
		t.Run("Agent: "+agentName, func(t *testing.T) {
			// Test that class info is not empty
			if classInfo.class == "" {
				t.Errorf("Agent %q has empty class", agentName)
			}
			if classInfo.emoji == "" {
				t.Errorf("Agent %q has empty emoji", agentName)
			}
			if classInfo.title == "" {
				t.Errorf("Agent %q has empty title", agentName)
			}
			
			// Test that classFor returns the same values
			gotClass, gotEmoji, gotTitle := classFor(agentName)
			if gotClass != classInfo.class {
				t.Errorf("classFor(%q) class = %q, want %q", agentName, gotClass, classInfo.class)
			}
			if gotEmoji != classInfo.emoji {
				t.Errorf("classFor(%q) emoji = %q, want %q", agentName, gotEmoji, classInfo.emoji)
			}
			if gotTitle != classInfo.title {
				t.Errorf("classFor(%q) title = %q, want %q", agentName, gotTitle, classInfo.title)
			}
		})
	}
}

func TestClassDistribution(t *testing.T) {
	// Test that we have a good distribution of classes
	classCount := make(map[string]int)
	
	for _, classInfo := range agentClasses {
		classCount[classInfo.class]++
	}
	
	// Verify we have multiple unique classes
	if len(classCount) < 5 {
		t.Errorf("Expected at least 5 different classes, got %d: %v", len(classCount), classCount)
	}
	
	// Log the distribution for manual inspection
	t.Logf("Class distribution: %v", classCount)
	
	// Verify specific expected classes exist
	expectedClasses := []string{"Overlord", "Engineer", "Artificer", "Bard", "Ranger", "Scribe", "Shadow", "Smith", "Gladiator", "Scout", "Sovereign", "Cleric", "Jester", "Sage", "Tinker", "Sentinel"}
	
	for _, expectedClass := range expectedClasses {
		if classCount[expectedClass] == 0 {
			t.Errorf("Expected class %q not found in agent classes", expectedClass)
		}
	}
}

func BenchmarkLevelForXP(b *testing.B) {
	// Benchmark the XP calculation function
	testXPs := []int{0, 100, 1000, 10000, 100000, 500000}
	
	for _, xp := range testXPs {
		b.Run(fmt.Sprintf("XP_%d", xp), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				levelForXP(xp)
			}
		})
	}
}

func BenchmarkClassFor(b *testing.B) {
	// Benchmark the class assignment function
	testAgents := []string{"claxon", "fionn", "main", "unknown-agent", "MIXED_CASE_AGENT"}
	
	for _, agent := range testAgents {
		b.Run(agent, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				classFor(agent)
			}
		})
	}
}