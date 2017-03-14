package models

import (
	"testing"
)

func Test_buildWordMap(t *testing.T) {

	var tests = []struct {
		input string
		want  int
	}{
		{"foo\n", 1},
		{"foo", 1},
		{"Call me Ishmael.\n", 3},
		{"Call me Ishmael.", 3},
		{"Hodor, hodor hodor. Hodor! Hodor hodor hodor hodor hodor hodor.\n", 1},
		{"Hodor, hodor hodor. Hodor! Hodor hodor hodor hodor hodor hodor.\n", 1},
		{"  the space test  ", 3},
		{"!   # %  the space test  !!!. \n", 3},
		{"!   # %  .the space test agian* !!!. \n", 4},
		{".....this is real-time dash_test ", 4},
		{"It was the best of times, it was the worst of times, it was the age of wisdom, it was the age of foolishness, it was the epoch of belief, it was the epoch of incredulity, it was the season of Light, it was the season of Darkness, it was the spring of hope, it was the winter of despair, we had everything before us, we had nothing before us, we were all going direct to Heaven, we were all going direct the other way - in short, the period was so far like the present period, that some of its noisiest authorities insisted on its being received, for good or for evil, in the superlative degree of comparison only.", 58},
		{"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.", 63},
	}

	for _, test := range tests {
		trollCap := &TrollCaptcha{
			Index: 0,
			Text:  test.input,
		}
		trollCap.buildWordMap(test.input)
		wordsSize := len(trollCap.Words)
		if wordsSize != test.want {
			t.Errorf("buildWordMap(%q) = got %v - want %v slice size", test.input, wordsSize, test.want)
		}

		mapSize := len(trollCap.WordMap)
		if mapSize != test.want {
			t.Errorf("buildWordMap(%q) = got %v - want %v map size", test.input, mapSize, test.want)
		}
	}
}

func Test_buildExclusionList(t *testing.T) {

	var tests = []struct {
		input string
		want  int
	}{
		{"foo\n", 0},
		{"foo", 0},
		{"Call me Ishmael. Hodor Hodor, test fun\n", 3},
		{"Call me Ishmael.", 2},
		{"Hodor, hodor hodor. Hodor! Hodor hodor hodor hodor hodor hodor.\n", 0},
		{"Hodor, hodor hodor. Hodor! Hodor hodor hodor hodor hodor hodor.\n", 0},
		{"  the space test  ", 2},
		{"the space test may the trolls be here", 3},
	}

	for _, test := range tests {
		trollCap := &TrollCaptcha{
			Index: 0,
			Text:  test.input,
		}
		trollCap.buildWordMap(test.input)
		trollCap.buildExclusionList(false, 3)
		exclusionsSize := len(trollCap.Exclusions)
		if exclusionsSize != test.want {
			t.Errorf("buildExclusionList(%q) = got %v - want %v exclusion size", test.input, exclusionsSize, test.want)
		}
	}
}

func Test_ValidateClientCaptcha(t *testing.T) {
	var tests = []struct {
		input       string
		clientInput []ClientWord
		exclusions  []string
		want        bool
	}{
		{input: "foo", clientInput: []ClientWord{{"foo", 1}}, want: true},
		{input: "foo bar test here", clientInput: []ClientWord{{"foo", 1}, {"bar", 1}, {"test", 0}, {"here", 0}}, exclusions: []string{"test", "here"}, want: true},
		{input: "a b c d", clientInput: []ClientWord{{"a", 1}, {"b", 1}, {"c", 0}, {"d", 1}}, exclusions: []string{"c"}, want: true},
		{input: "foo bar test here", clientInput: []ClientWord{{"foo", 1}, {"bar", 1}, {"test", 1}, {"here", 0}}, exclusions: []string{"test", "here"}, want: false},
		{input: "foo bar test here", clientInput: []ClientWord{{"foo", 2}, {"bar", 0}, {"test", 0}, {"here", 0}}, exclusions: []string{"test", "here"}, want: false},
		{input: ".....foo foo bar test test one here here \n", clientInput: []ClientWord{{"foo", 2}, {"bar", 0}, {"test", 0}, {"one", 0}, {"here", 2}}, exclusions: []string{"bar", "test", "one"}, want: true},
		{input: ".....foo foo bar test test one here here \n", clientInput: []ClientWord{{"foo", 2}, {"bar", 0}, {"test", 0}, {"one", 1}, {"here", 2}}, exclusions: []string{"bar", "test", "one"}, want: false},
	}

	for _, test := range tests {
		trollCap := &TrollCaptcha{
			Index: 0,
			Text:  test.input,
		}
		trollCap.Id = buildStringId(test.input)
		trollCap.buildWordMap(test.input)

		testCap := &TrollCaptcha{}
		testCap.Text = test.input
		testCap.ClientWords = test.clientInput

		//Given randomness of normal operation need to run this manually
		//that way wordmap is cleared in non random way
		if len(trollCap.Words) > 1 {
			trollCap.Exclusions = test.exclusions
			for _, v := range trollCap.Exclusions {
				trollCap.WordMap[v] = 0
			}
		}
		if s, got := trollCap.ValidateClientCaptcha(testCap); got != test.want {
			t.Errorf("ValidateClientCaptcha(%q) = got %v wanted %v ---- %v", test.clientInput, got, test.want, s)
		}
	}
}
