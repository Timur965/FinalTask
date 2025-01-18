package handlerequest

import (
	"os"
	"strings"
)

type CensorShip struct {
	forbiddenWords map[string]string
}

func newCensorShip() (*CensorShip, error) {
	var cs CensorShip
	textByte, err := os.ReadFile("forbiddenWords.txt")
	if err != nil {
		return nil, err
	}
	textStr := string(textByte)
	cs.forbiddenWords = make(map[string]string)
	for _, word := range strings.Split(textStr, "\n") {
		word = strings.TrimSpace(word)
		cs.forbiddenWords[word] = strings.Repeat("*", len(word))
	}

	return &cs, nil
}

func (cs *CensorShip) IsForbiddenWords(text string) string {
	result := text
	for key, value := range cs.forbiddenWords {
		result = strings.ReplaceAll(result, key, value)
	}
	return result
}
