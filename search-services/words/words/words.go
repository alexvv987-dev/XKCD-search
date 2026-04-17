package words

import (
	"log/slog"
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
	"github.com/kljensen/snowball/english"
)

func Norm(phrase string) []string {
	words := strings.FieldsFunc(phrase, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	stemmed := make(map[string]struct{})
	for _, word := range words {
		stemmedWord, err := snowball.Stem(word, "english", true)
		if err != nil {
			slog.Error("stem failed", "error", err)
			continue
		}
		_, ok := stemmed[stemmedWord]
		if english.IsStopWord(stemmedWord) || ok {
			continue
		}
		stemmed[stemmedWord] = struct{}{}
	}
	return slices.Collect(maps.Keys(stemmed))
}
