package quest

import (
	"errors"
	"strings"
	"unicode"
)

// ErrWrongAnswer is returned when a quest with a definite answer was not
// answered correctly. The text is a translation key for the UI.
var ErrWrongAnswer = errors.New("quest.wrongAnswer")

// CheckAnswers verifies answers for quests that have a right answer:
// quizzes and answer quests with a check. Other quests always pass.
func CheckAnswers(q Quest, answers []string) error {
	in := q.Interaction
	switch {
	case in.Kind == KindQuiz:
		if len(answers) == 0 || !isOption(q, in.Answer-1, answers[0]) {
			return ErrWrongAnswer
		}
	case in.Kind == KindAnswer && in.Check != "":
		if len(answers) != len(in.Prompts) {
			return ErrWrongAnswer
		}
		for i := range answers {
			if !AnswerOK(q, answers, i) {
				return ErrWrongAnswer
			}
		}
	}
	return nil
}

// AnswerOK reports whether answer i is right, given all the answers. Empty
// answers are never right for a checked quest.
func AnswerOK(q Quest, answers []string, i int) bool {
	if i < 0 || i >= len(answers) || strings.TrimSpace(answers[i]) == "" {
		return false
	}
	a := answers[i]
	switch q.Interaction.Check {
	case CheckInitials:
		return firstLetter(a) == firstLetter(q.Interaction.Prompts[i])
	case CheckChain:
		if i == 0 {
			return firstLetter(a) != 0
		}
		prev := lastLetter(answers[i-1])
		return prev != 0 && firstLetter(a) == prev
	case CheckReverse:
		if i == 0 {
			return len(letters(a)) >= 2
		}
		return letters(a) == reverse(letters(answers[0])) && len(letters(a)) >= 2
	}
	return true
}

// isOption reports whether text is option i of q, in any language.
func isOption(q Quest, i int, text string) bool {
	if i < 0 || i >= len(q.Interaction.Prompts) {
		return false
	}
	same := func(s string) bool { return strings.EqualFold(strings.TrimSpace(s), strings.TrimSpace(text)) }
	if same(q.Interaction.Prompts[i]) {
		return true
	}
	for _, tr := range q.I18n {
		if i < len(tr.Prompts) && same(tr.Prompts[i]) {
			return true
		}
	}
	return false
}

// letters keeps only the letters and digits of s, lower-cased.
func letters(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func firstLetter(s string) rune {
	for _, r := range letters(s) {
		return r
	}
	return 0
}

func lastLetter(s string) rune {
	l := []rune(letters(s))
	if len(l) == 0 {
		return 0
	}
	return l[len(l)-1]
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
