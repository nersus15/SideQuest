package quest

import (
	"errors"
	"fmt"
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
	// Two quests a day for a year.
	if len(qs) < 730 {
		t.Fatalf("expected at least 730 builtin quests, got %d", len(qs))
	}
	perCat := map[Category]int{}
	perDifficulty := map[int]int{}
	for _, q := range qs {
		perCat[q.Category]++
		perDifficulty[q.Difficulty]++
		if q.Source != SourceBuiltin || !strings.HasPrefix(q.Pack, "builtin-") {
			t.Fatalf("quest %s has source %q pack %q", q.ID, q.Source, q.Pack)
		}
	}
	for _, c := range Categories {
		if perCat[c] < 100 {
			t.Errorf("category %s has only %d quests", c, perCat[c])
		}
	}
	// New players only see easy quests, so most quests must be easy.
	if perDifficulty[1]*2 < len(qs) {
		t.Errorf("only %d of %d quests are easy", perDifficulty[1], len(qs))
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

func TestBuiltinQuestsAreTranslated(t *testing.T) {
	qs, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range qs {
		tr, ok := q.I18n["id"]
		if !ok || tr.Title == "" || tr.Description == "" {
			t.Errorf("%s: missing Indonesian title or description", q.ID)
			continue
		}
		for _, tw := range q.Twists {
			if tr.Twists[tw.ID] == "" {
				t.Errorf("%s: twist %s is not translated", q.ID, tw.ID)
			}
		}
	}
}

func TestInteractionValidation(t *testing.T) {
	base := `{"format":1,"id":"p1","name":"P","quests":[{"id":"q1","title":"T","description":"D","category":"chill","minutes":1,"difficulty":1,"xp":10,"interaction":%s%s}]}`
	ok := map[string]string{
		"answer":    `{"kind":"answer","prompts":["A","B"]}`,
		"checklist": `{"kind":"checklist","prompts":["A","B"]}`,
		"timer":     `{"kind":"timer","seconds":30,"guide":"breath"}`,
		"counter":   `{"kind":"counter","target":5}`,
		"doodle":    `{"kind":"doodle"}`,
	}
	for name, in := range ok {
		if _, err := ParsePack(strings.NewReader(fmt.Sprintf(base, in, "")), SourceImported); err != nil {
			t.Errorf("%s should be valid: %v", name, err)
		}
	}
	bad := map[string]string{
		"unknown kind":       `{"kind":"shell","prompts":["rm"]}`,
		"answer no prompts":  `{"kind":"answer"}`,
		"too many answers":   `{"kind":"answer","prompts":["1","2","3","4","5","6"]}`,
		"checklist one step": `{"kind":"checklist","prompts":["A"]}`,
		"timer too long":     `{"kind":"timer","seconds":3600}`,
		"timer bad guide":    `{"kind":"timer","seconds":10,"guide":"hypnosis"}`,
		"counter huge":       `{"kind":"counter","target":1000}`,
		"doodle with prompt": `{"kind":"doodle","prompts":["x"]}`,
		"seconds on counter": `{"kind":"counter","seconds":10}`,
	}
	for name, in := range bad {
		if _, err := ParsePack(strings.NewReader(fmt.Sprintf(base, in, "")), SourceImported); err == nil {
			t.Errorf("%s should be rejected", name)
		}
	}
	// Translations must match the prompts and twists they translate.
	badI18n := map[string]string{
		"unknown language": `,"i18n":{"xx":{"title":"T"}}`,
		"prompt count":     `,"i18n":{"id":{"prompts":["satu"]}}`,
		"unknown twist":    `,"i18n":{"id":{"twists":{"nope":"x"}}}`,
	}
	for name, extra := range badI18n {
		if _, err := ParsePack(strings.NewReader(fmt.Sprintf(base, ok["answer"], extra)), SourceImported); err == nil {
			t.Errorf("%s should be rejected", name)
		}
	}
}

func TestPickRespectsDifficultyUnlock(t *testing.T) {
	qs := []Quest{
		{ID: "easy", Difficulty: 1, Repeatable: true},
		{ID: "medium", Difficulty: 2, Repeatable: true},
		{ID: "bold", Difficulty: 3, Repeatable: true},
		{ID: "mine", Difficulty: 3, Repeatable: true, Source: SourceCustom},
	}
	rng := testRand()
	seen := map[string]bool{}
	for i := 0; i < 300; i++ {
		o, _ := Pick(qs, Prefs{MaxDifficulty: 1}, Memory{}, t0, "", rng)
		seen[o.Quest.ID] = true
	}
	if seen["medium"] || seen["bold"] || !seen["easy"] || !seen["mine"] {
		t.Fatalf("difficulty lock not applied: %v", seen)
	}
}

func TestRatingsTuneSelection(t *testing.T) {
	qs := []Quest{
		{ID: "liked", Category: Chill, Repeatable: true},
		{ID: "plain", Category: Chill, Repeatable: true},
		{ID: "disliked", Category: Chill, Repeatable: true},
	}
	mem := Memory{Ratings: map[string]Rating{"liked": RatingUp, "disliked": RatingDown}}
	rng := testRand()
	seen := map[string]int{}
	for i := 0; i < 3000; i++ {
		o, _ := Pick(qs, Prefs{}, mem, t0, "", rng)
		seen[o.Quest.ID]++
	}
	if !(seen["liked"] > seen["plain"] && seen["plain"] > seen["disliked"]) {
		t.Fatalf("expected liked > plain > disliked, got %v", seen)
	}
	if seen["disliked"] == 0 {
		t.Fatal("a disliked quest should still come up now and then")
	}
}

func TestDislikedQuestIsStillOfferedWhenAlone(t *testing.T) {
	qs := []Quest{{ID: "only", Repeatable: true}}
	mem := Memory{Ratings: map[string]Rating{"only": RatingDown}}
	if o, err := Pick(qs, Prefs{}, mem, t0, "", testRand()); err != nil || o.Quest.ID != "only" {
		t.Fatalf("got %v %v", o.Quest.ID, err)
	}
}

func TestCategoryAffinityIsGentleAndBounded(t *testing.T) {
	var qs []Quest
	ratings := map[string]Rating{}
	for i := 0; i < 20; i++ {
		up := Quest{ID: fmt.Sprintf("p%d", i), Category: Play}
		down := Quest{ID: fmt.Sprintf("s%d", i), Category: Social}
		qs = append(qs, up, down)
		ratings[up.ID] = RatingUp
		ratings[down.ID] = RatingDown
	}
	qs = append(qs, Quest{ID: "x", Category: Explore})
	ratings["ghost"] = RatingUp // a quest that no longer exists is ignored

	a := categoryAffinity(qs, ratings)
	if a[Play] > 1.3 || a[Play] < 1.25 || a[Social] < 0.8 || a[Social] > 0.85 || a[Explore] != 1 {
		t.Fatalf("unexpected affinity: %v", a)
	}
	one := categoryAffinity(qs, map[string]Rating{"p0": RatingUp})
	if one[Play] > 1.15 {
		t.Fatalf("a single like should only nudge a category, got %v", one[Play])
	}
	if categoryAffinity(qs, nil) != nil {
		t.Fatal("no ratings means no affinity")
	}
}
