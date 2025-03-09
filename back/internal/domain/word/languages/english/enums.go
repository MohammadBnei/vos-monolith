package english

type WordType string

const (
	Noun      WordType = "noun"
	Verb      WordType = "verb"
	Adjective WordType = "adjective"
	Adverb    WordType = "adverb"
	Pronoun   WordType = "pronoun"
	Preposition WordType = "preposition"
	Conjunction WordType = "conjunction"
	Interjection WordType = "interjection"
)

func IsValidWordType(wt WordType) bool {
	switch wt {
	case Noun, Verb, Adjective, Adverb, Pronoun, Preposition, Conjunction, Interjection:
		return true
	default:
		return false
	}
}
