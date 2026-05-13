package words_test

import (
	"sort"
	"testing"

	"github.com/kljensen/snowball"
	"github.com/stretchr/testify/assert"
	"yadro.com/course/words/words"
)

func TestNorm(t *testing.T) {
	stem := func(word string) string {
		s, err := snowball.Stem(word, "english", true)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty_string",
			input: "",
			want:  []string{},
		},
		{
			name:  "stop_words_only",
			input: "the a an",
			want:  []string{},
		},
		{
			name:  "duplicate_forms_collapse",
			input: "runs running",
			want:  []string{stem("running")},
		},
		{
			name:  "punctuation_as_separator",
			input: "cats,dogs",
			want:  []string{stem("cats"), stem("dogs")},
		},
		{
			name:  "digits_not_filtered",
			input: "linux 2",
			want:  []string{stem("linux"), "2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := words.Norm(tt.input)
			if len(tt.want) == 0 {
				assert.Empty(t, got)
			} else {
				sort.Strings(got)
				sort.Strings(tt.want)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
