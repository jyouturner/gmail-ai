package nlp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jdkato/prose/v3"
)

func ExtractTopSentenseFrom(n int, text string) (string, error) {

	doc, err := prose.NewDocument(text)
	if err != nil {
		fmt.Println("Error while creating document:", err)
		return "", err
	}

	var sentences []string
	for _, s := range doc.Sentences() {
		sentences = append(sentences, s.Text)
	}

	// Sort the sentences by length
	sort.Slice(sentences, func(i, j int) bool {
		return len(sentences[i]) > len(sentences[j])
	})

	// Extract the top N sentences
	topSentences := sentences[:n]

	// Join the top sentences and print
	topText := strings.Join(topSentences, " ")
	fmt.Println("Top text:", topText)
	return topText, nil
}
