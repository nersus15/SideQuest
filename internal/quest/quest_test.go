package quest

import (
	"errors"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testRand() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

var t0 = time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)

func TestBuiltinPackIsValid(t *testing.T) {
	qs, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) < 40 || len(qs) > 60 {
		t.Fatalf("expected 40-60 builtin quests, got %d", len(qs))
	}
	perCat := map[Category]int{}
	for _, q := range qs {
		perCat[q.Category]++
		if q.Source != SourceBuiltin || q.Pack != "builtin" {
			t.Fatalf("quest %s has source %q pack %q", q.ID, q.Source, q.Pack)
		}
	}
	for _, c := range Categories {
		if perCat[c] < 5 {
			t.Errorf("category %s has only %d quests", c, perCat[c])
		}
	}
}

func TestParsePackRejectsBadData(t *testing.T) {
	cases := map[string]string{
		"unknown field": `{"format":1,"id":"p1","name":"P","quests":[{"id":"q1","title":"T","description":"D","category":"chill","minutes":1,"difficulty":1,"xp":10,"command":"rm -rf /"}]}`,
		"bad category":  `{"format":1,"id":"p1","name":"P","quests":[{"id":"q1","title":"T","description":"D","category":"hacking","minutes":1,"difficulty":1,"xp":10}]}`,
		"xp too high":   `{"format":1,"id":"p1","name":"P","quests":[{"id":"q1","title":"T","description":"D","category":"chill","minutes":1,"difficulty":1,"xp":9999}]}`,
		"bad id":        `{"format":1,"id":"p1","name":"P","quests":[{"id":"../evil","title":"T","description":"D","category":"chill","minutes":1,"difficulty":1,"xp":10}]}`,
		"duplicate id":  `{"format":1,"id":"p1","name":"P","quests":[{"id":"q1","title":"T","description":"D","category":"chill","minutes":1,"difficulty":1,"xp":10},{"id":"q1","title":"T","description":"D","category":"chill","minutes":1,"difficulty":1,"xp":10}]}`,
		"wrong format":  `{"format":2,"id":"p1","name":"P","quests":[{"id":"q1","title":"T","description":"D","category":"chill","minutes":1,"difficulty":1,"xp":10}]}`,
		"empty":         `{"format":1,"id":"p1","name":"P","quests":[]}`,
		"not json":      `<script>alert(1)</script>`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParsePack(strings.NewReader(body), SourceImported); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestParsePackOverridesSource(t *testing.T) {
	body := `{"format":1,"id":"mine","name":"Mine","quests":[{"id":"q1","title":" Hi ","description":"Wave","category":"social","minutes":1,"difficulty":1,"xp":10,"source":"builtin"}]}`
	p, err := ParsePack(strings.NewReader(body), SourceCustom)
	if err != nil {
		t.Fatal(err)
	}
	q := p.Quests[0]
	if q.Source != SourceCustom || q.Pack != "mine" || q.Title != "Hi" {
		t.Fatalf("unexpected normalised quest: %+v", q)
	}
}

func TestParsePackRejectsHugeInput(t *testing.T) {
	body := `{"format":1,"id":"p1","name":"` + strings.Repeat("a", MaxPackBytes) + `"}`
	_, err := ParsePack(strings.NewReader(body), SourceImported)
	if err == nil || !strings.Contains(err.Error(), "larger") {
		t.Fatalf("expected size error, got %v", err)
	}
}

func TestMergeKeepsFirst(t *testing.T) {
	a := []Quest{{ID: "x", Title: "builtin"}}
	b := []Quest{{ID: "x", Title: "imported"}, {ID: "y"}}
	got := Merge(a, b)
	if len(got) != 2 || got[0].Title != "builtin" {
		t.Fatalf("unexpected merge: %+v", got)
	}
}

func sample(n int) []Quest {
	var qs []Quest
	cats := Categories
	for i := 0; i < n; i++ {
		qs = append(qs, Quest{ID: string(rune('a'+i)) + "q", Category: cats[i%len(cats)], XP: 10, Repeatable: true})
	}
	return qs
}

func TestPickNeverRepeatsExcludedWhenAlternativesExist(t *testing.T) {
	qs := sample(3)
	rng := testRand()
	for i := 0; i < 200; i++ {
		o, err := Pick(qs, Prefs{}, Memory{}, t0, "aq", rng)
		if err != nil {
			t.Fatal(err)
		}
		if o.Quest.ID == "aq" {
			t.Fatal("picked excluded quest")
		}
	}
}

func TestPickAvoidsRecentlyShown(t *testing.T) {
	qs := sample(10)
	mem := Memory{RecentlyShown: []string{"aq", "bq", "cq", "dq", "eq"}}
	rng := testRand()
	for i := 0; i < 200; i++ {
		o, _ := Pick(qs, Prefs{}, mem, t0, "", rng)
		if strings.Contains("aq bq cq dq eq", o.Quest.ID) {
			t.Fatalf("picked recently shown quest %s", o.Quest.ID)
		}
	}
}

func TestPickRespectsCooldownAndOneOffs(t *testing.T) {
	qs := []Quest{
		{ID: "rest", Repeatable: true, CooldownHours: 2},
		{ID: "once", Repeatable: false},
		{ID: "free", Repeatable: true},
	}
	mem := Memory{LastCompleted: map[string]time.Time{
		"rest": t0.Add(-1 * time.Hour),
		"once": t0.Add(-1000 * time.Hour),
	}}
	rng := testRand()
	for i := 0; i < 100; i++ {
		o, _ := Pick(qs, Prefs{}, mem, t0, "", rng)
		if o.Quest.ID != "free" {
			t.Fatalf("expected only 'free', got %s", o.Quest.ID)
		}
	}
	// After the cooldown, "rest" may return; "once" never does.
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		o, _ := Pick(qs, Prefs{}, mem, t0.Add(3*time.Hour), "", rng)
		seen[o.Quest.ID] = true
	}
	if !seen["rest"] || seen["once"] {
		t.Fatalf("unexpected availability after cooldown: %v", seen)
	}
}

func TestPickHonoursPrefsButFallsBack(t *testing.T) {
	qs := []Quest{
		{ID: "talk", Category: Social, Needs: []Need{NeedPeople}, Repeatable: true},
		{ID: "doodle", Category: Creative, Needs: []Need{NeedPaper}, Repeatable: true},
		{ID: "breathe", Category: Chill, Repeatable: true},
	}
	prefs := Prefs{Muted: []Category{Chill}, Avoid: []Need{NeedPeople}}
	rng := testRand()
	for i := 0; i < 100; i++ {
		o, _ := Pick(qs, prefs, Memory{}, t0, "", rng)
		if o.Quest.ID != "doodle" {
			t.Fatalf("prefs ignored: got %s", o.Quest.ID)
		}
	}
	// When prefs exclude everything else, the player still gets a quest.
	o, err := Pick(qs, prefs, Memory{}, t0, "doodle", rng)
	if err != nil || o.Quest.ID == "doodle" {
		t.Fatalf("expected fallback quest, got %v %v", o.Quest.ID, err)
	}
}

func TestPickSingleQuestReturnsItEvenIfExcluded(t *testing.T) {
	qs := []Quest{{ID: "only"}}
	o, err := Pick(qs, Prefs{}, Memory{}, t0, "only", testRand())
	if err != nil || o.Quest.ID != "only" {
		t.Fatalf("got %v %v", o.Quest.ID, err)
	}
	if _, err := Pick(nil, Prefs{}, Memory{}, t0, "", testRand()); !errors.Is(err, ErrNoQuests) {
		t.Fatalf("expected ErrNoQuests, got %v", err)
	}
}

func TestPickPrefersVariety(t *testing.T) {
	qs := sample(12)
	rng := testRand()
	mem := Memory{LastCategory: Creative}
	same := 0
	for i := 0; i < 1000; i++ {
		o, _ := Pick(qs, Prefs{}, mem, t0, "", rng)
		if o.Quest.Category == Creative {
			same++
		}
	}
	// 2 of 12 are creative; unweighted that is ~167/1000. Weighted should be lower.
	if same > 140 {
		t.Fatalf("expected variety nudge, creative picked %d/1000", same)
	}
}

func TestOfferSuggestsTwist(t *testing.T) {
	qs := []Quest{{ID: "tw", Twists: []Twist{{ID: "a", Text: "A"}, {ID: "b", Text: "B"}}}}
	o, _ := Pick(qs, Prefs{}, Memory{}, t0, "", testRand())
	if o.Twist == nil {
		t.Fatal("expected a twist suggestion")
	}
}

func TestLoadDirSkipsInvalidPacks(t *testing.T) {
	dir := t.TempDir()
	good := `{"format":1,"id":"good","name":"Good","quests":[{"id":"g1","title":"Hi","description":"Wave","category":"social","minutes":1,"difficulty":1,"xp":10}]}`
	if err := os.WriteFile(filepath.Join(dir, "good.json"), []byte(good), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte(`{"format":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignored"), 0o644); err != nil {
		t.Fatal(err)
	}
	qs, problems := LoadDir(dir, SourceImported)
	if len(qs) != 1 || qs[0].ID != "g1" || qs[0].Source != SourceImported {
		t.Fatalf("unexpected quests: %+v", qs)
	}
	if len(problems) != 1 || !strings.Contains(problems[0].Error(), "bad.json") {
		t.Fatalf("expected one problem for bad.json, got %v", problems)
	}
	if qs, problems := LoadDir(filepath.Join(dir, "missing"), SourceImported); len(qs)+len(problems) != 0 {
		t.Fatal("missing dir should yield nothing")
	}
}

func TestClockMovedBackDoesNotLockQuests(t *testing.T) {
	qs := []Quest{{ID: "a", Repeatable: true}, {ID: "b", Repeatable: true}}
	// Completed "in the future": the user moved the clock back a day.
	mem := Memory{
		LastCompleted: map[string]time.Time{"a": t0.Add(24 * time.Hour)},
		LastSkipped:   map[string]time.Time{"b": t0.Add(24 * time.Hour)},
	}
	seen := map[string]bool{}
	rng := testRand()
	for i := 0; i < 100; i++ {
		o, _ := Pick(qs, Prefs{}, mem, t0, "", rng)
		seen[o.Quest.ID] = true
	}
	if !seen["a"] || !seen["b"] {
		t.Fatalf("future timestamps should not rest quests: %v", seen)
	}
}
