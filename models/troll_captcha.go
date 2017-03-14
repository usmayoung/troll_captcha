package models

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/gorilla/schema"
	"math/rand"
	"strings"
	"time"
	"unicode"
)

var decoder *schema.Decoder
var count int

func init() {
	decoder = schema.NewDecoder()
}

//Main data structure for TrollCaptcha.  This object is cached on init() from main
type TrollCaptcha struct {
	Id      string         //md5 hash of text, used to uniquely identify a troll captcha, also used to verify text sent by client
	Index   int            //index position in TrollCaptcha cache
	WordMap map[string]int //key-value map of every unique key word and its word count within TrollCaptcha,
	// find in O(1) on average, value of 0 indicates exclusion
	Words       []string     //slice of all unique words, mostly used just used for easy rendering on client
	Exclusions  []string     //slice of all excluded words, mostly used just used for easy rendering on client
	Text        string       `schema:"text"`        //original text of captcha, also used to parse text post from client
	ClientWords []ClientWord `schema:"ClientWords"` //slice of all text counts posted by client
}

//Object for collecting client's post of key words and accociated word counts
type ClientWord struct {
	Word  string //word
	Count int    //count indicating how many times the word appeared in word, 0 or " " indicates it is ignored
}

type Message struct {
	Id int
	Type string `json:"type"`
	Value Value
}

type Value struct {
	Joke string `json:"joke"`
}

//Constructor for TrollCaptcha, calls t.buildWordMap in order to trigger
//the creation of TrollCaptcha fields
func NewTrollCaptcha(text string, index int) *TrollCaptcha {
	trollCap := &TrollCaptcha{
		Index: index,
		Text:  text,
	}
	trollCap.buildWordMap(text)
	trollCap.Id = buildStringId(text)
	trollCap.buildExclusionList(true, 0)
	return trollCap
}

//buildWordMap is linear O(n), iterates over each rune in input text to determine
//where to split the text into words.  Word doesn't start processing
//until first letter, if there is a mark token within a word, "-" for example
//the mark will be included as part of the word.  Once space is found the word is added
//to a map, with the word as the key and increments the count for the value.
//By using a map the keys (words), must be unique, and therefore searching
//and inserting increments in constant time on average
func (t *TrollCaptcha) buildWordMap(text string) {

	wordMap := make(map[string]int)
	words := make([]string, 0)

	buildString := ""
	processingWord := false
	for _, character := range text {
		if unicode.IsLetter(character) {
			buildString += string(unicode.ToLower(character))
			processingWord = true
		} else if unicode.IsMark(character) {
			if processingWord {
				buildString += string(character)
			}
		} else if unicode.IsSpace(character) {
			if processingWord {
				addStringToWords(&wordMap, &words, buildString)
				buildString = ""
				processingWord = false
			}
		}

	}
	//make sure word is not finished processing, could happen without a string ending
	//with \n
	if buildString != "" {
		addStringToWords(&wordMap, &words, buildString)
	}
	//word map now contains keys for each unique word and value with the repeated count of the word
	t.WordMap = wordMap
	//words are unique and striped of punctuation on either end of the word at this point
	t.Words = words

}

//helper function to add text word to map and slice, pass value of reference
func addStringToWords(wMap *map[string]int, wSlice *[]string, text string) {
	(*wMap)[text]++
	if v, _ := (*wMap)[text]; v == 1 {
		*wSlice = append(*wSlice, text)
	}
}

//buildRandomExclusionList is used to build the excluded list, if the text has only one unique
//element, then the exclusion list is nil.  Otherwise, the exclusion is a random size
//from 1 to to the size of unique words, but not exceeding 5, the wordmap is updated
// to reflect a 0 for ignored words
//THE EXCLUDED LIST WILL BE RANDOM NO MATTER WHAT INPUT
// BECAUSE RANGE OVER T.WORDMAP IS RANDOM ORDER
func (t *TrollCaptcha) buildExclusionList(random bool, size int) {
	uniqueWordCount := len(t.WordMap)

	if uniqueWordCount <= 1 {
		t.Exclusions = nil
	} else {
		//This random is only to determine the size of exclusion list,
		//not the randomness of the elements in the list
		if random {
			s := rand.NewSource(time.Now().Unix())
			r := rand.New(s)
			randSourceSize := r.Intn(uniqueWordCount) + 1
			randSize := Min(randSourceSize, 5)

			size = randSize
		}

		if size == uniqueWordCount {
			size--
		}

		excludedList := make([]string, size)

		count := 0

		//this range will be random, therefore, excludedList will
		//be unique and random
		for key := range t.WordMap {
			if count < size {
				excludedList[count] = key
				t.WordMap[key] = 0
				count++

			}
		}
		t.Exclusions = excludedList
	}
}

//buildStringId takes input text and md5 hashes in order to create unique id for the captcha
//whitespace is always trimed before creation
func buildStringId(input string) string {
	trimmed := strings.TrimSpace(input)
	hash := md5.Sum([]byte(trimmed))
	return hex.EncodeToString(hash[:])
}

//ValidateClientCaptcha validates the valid cached trollcaptcha with the input provided by client
//first checks if the text string submited matches the correct original text
//second checks if the word list from client is the correct size as the
//original unique word on the server
//if all that passes, then interate over each client input word, check if the
//word exists and the count is correct, a count of " " or 0 is an ignored word
//each search for word counts in constant time on averate, but linear if all
//words need to be searched
func (t *TrollCaptcha) ValidateClientCaptcha(c *TrollCaptcha) (string, bool) {
	if buildStringId(c.Text) != t.Id {
		return "Sorry Troll but your text is not valid!", false
	}
	if len(t.Words) != len(c.ClientWords) {
		return "Sorry Troll but you didn't send all your words", false
	}

	for _, w := range c.ClientWords {
		if count, ok := t.WordMap[w.Word]; !ok || count != w.Count {
			return "Sorry Troll but you can't add!", false
		}
	}
	return "Success", true

}

//Helper function to determine min value between two ints
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
