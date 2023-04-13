package nlp

import (
	"math"
	"sort"
	"strings"

	"github.com/jdkato/prose/v3"
	"github.com/jyouturer/gmail-ai/internal/logging"
	"go.uber.org/zap"
)

func ExtractTopSentenseFrom(n int, text string) (string, error) {

	doc, err := prose.NewDocument(text)
	if err != nil {
		logging.Logger.Info("Error while creating document:", zap.Error(err))
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
	topSentences := sentences[:int(math.Min(float64(n), float64(len(sentences))))]

	// Join the top sentences and print
	topText := strings.Join(topSentences, " ")
	logging.Logger.Info("Top text:", zap.String("top text", topText))
	return topText, nil
}
