package quest

import (
	"errors"
	"testing"
)

func TestQuizNeedsTheRightOption(t *testing.T) {
	q := Quest{
		Interaction: Interaction{Kind: KindQuiz, Prompts: []string{"33", "43", "53"}, Answer: 2},
		I18n:        map[string]Translation{"id": {Prompts: []string{"tiga tiga", "empat tiga", "lima tiga"}}},
	}
	if err := CheckAnswers(q, []string{"43"}); err != nil {
		t.Fatalf("right answer rejected: %v", err)
	}
	if err := CheckAnswers(q, []string{"empat tiga"}); err != nil {
		t.Fatalf("right answer in Indonesian rejected: %v", err)
	}
	for _, wrong := range [][]string{{"33"}, {"made up"}, {}} {
		if err := CheckAnswers(q, wrong); !errors.Is(err, ErrWrongAnswer) {
			t.Errorf("%v should be wrong, got %v", wrong, err)
		}
	}
}

func TestAnswerChecks(t *testing.T) {
	initials := Quest{Interaction: Interaction{Kind: KindAnswer, Prompts: []string{"A...", "B...", "C..."}, Check: CheckInitials}}
	chain := Quest{Interaction: Interaction{Kind: KindAnswer, Prompts: []string{"1", "2", "3"}, Check: CheckChain}}
	rev := Quest{Interaction: Interaction{Kind: KindAnswer, Prompts: []string{"Food", "Backwards"}, Check: CheckReverse}}
	cases := []struct {
		name    string
		q       Quest
		answers []string
		ok      bool
	}{
		{"initials ok", initials, []string{"apple", "Banana", " cherry"}, true},
		{"initials wrong letter", initials, []string{"apple", "kiwi", "cherry"}, false},
		{"initials empty", initials, []string{"apple", "", "cherry"}, false},
		{"chain ok", chain, []string{"cat", "tiger", "rabbit"}, true},
		{"chain ignores punctuation", chain, []string{"cat!", "Tiger.", "rabbit"}, true},
		{"chain broken", chain, []string{"cat", "dog", "goat"}, false},
		{"reverse ok", rev, []string{"Nasi Goreng", "gnerog isan"}, true},
		{"reverse wrong", rev, []string{"pizza", "pizza"}, false},
		{"reverse made up", rev, []string{"pizza", "asdf"}, false},
		{"wrong count", chain, []string{"cat"}, false},
	}
	for _, c := range cases {
		err := CheckAnswers(c.q, c.answers)
		if (err == nil) != c.ok {
			t.Errorf("%s: got %v", c.name, err)
		}
	}
	free := Quest{Interaction: Interaction{Kind: KindAnswer, Prompts: []string{"1"}}}
	if err := CheckAnswers(free, []string{"anything"}); err != nil {
		t.Fatalf("unchecked answers must pass: %v", err)
	}
}

func TestCheckValidation(t *testing.T) {
	bad := []Interaction{
		{Kind: KindChoice, Prompts: []string{"A", "B"}, Check: CheckChain},
		{Kind: KindAnswer, Prompts: []string{"1"}, Check: CheckChain},
		{Kind: KindAnswer, Prompts: []string{"1", "2", "3"}, Check: CheckReverse},
		{Kind: KindAnswer, Prompts: []string{"123"}, Check: CheckInitials},
		{Kind: KindAnswer, Prompts: []string{"A"}, Check: "magic"},
	}
	for i, in := range bad {
		if problems := validateInteraction(&in); len(problems) == 0 {
			t.Errorf("case %d should be rejected", i)
		}
	}
}
