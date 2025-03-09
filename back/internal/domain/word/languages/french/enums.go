package french

type WordType string

const (
	Noun         WordType = "nom"
	Verb         WordType = "verbe"
	Adjective    WordType = "adjectif"
	Adverb       WordType = "adverbe"
	Pronoun      WordType = "pronom"
	Preposition  WordType = "préposition"
	Conjunction  WordType = "conjonction"
	Interjection WordType = "interjection"
)

type Gender string

const (
	Masculine Gender = "masculin"
	Feminine  Gender = "féminin"
	Plural    Gender = "pluriel"
)

func IsValidWordType(wt WordType) bool {
	switch wt {
	case Noun, Verb, Adjective, Adverb, Pronoun, Preposition, Conjunction, Interjection:
		return true
	default:
		return false
	}
}

func IsValidGender(g Gender) bool {
	switch g {
	case Masculine, Feminine, Plural:
		return true
	default:
		return false
	}
}
