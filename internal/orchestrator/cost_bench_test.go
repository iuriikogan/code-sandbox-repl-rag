package orchestrator

import (
	"fmt"
	"testing"

	"githuob.com/iuriikogan/code-sandbox-repl-rag/internal/data"
)

// approximateTokenCount estimates tokens by dividing character count by 4.
// This is a standard industry rule-of-thumb for LLM tokenization.
func approximateTokenCount(text string) float64 {
	return float64(len(text)) / 4.0
}

// Hypothetical pricing rates (USD per 1 Million tokens) based on Gemini pricing guidelines.
const (
	PriceProInputPer1M  = 3.50  // gemini-2.5-pro approx input cost
	PriceEmbeddingPer1M = 0.02  // text-embedding-004 approx cost
)

var contextSizes = []int{10, 100, 1000, 10000} // Multipliers for context generation (from small to very large)

func BenchmarkCost_StandardRouting(b *testing.B) {
	for _, size := range contextSizes {
		b.Run(fmt.Sprintf("ContextSize_%d", size), func(b *testing.B) {
			dataset := data.GenerateContext(size)
			// Standard routing sends everything to Pro
			prompt := "Extract comprehensive summary of all FINANCE related events.\n\n" + dataset

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tokens := approximateTokenCount(prompt)
				cost := (tokens / 1000000.0) * PriceProInputPer1M

				b.ReportMetric(tokens, "tokens/op")
				b.ReportMetric(cost, "USD/op")
			}
		})
	}
}

func BenchmarkCost_RAGRouting(b *testing.B) {
	for _, size := range contextSizes {
		b.Run(fmt.Sprintf("ContextSize_%d", size), func(b *testing.B) {
			dataset := data.GenerateContext(size)
			
			// Simulate the chunk size: 
			// In RAG, regardless of total size, we only send the Top 5 chunks to the Pro model.
			// Base chunk is approx 300 chars, so 5 chunks = ~1500 chars.
			topChunksLen := 1500
			if len(dataset) < topChunksLen {
				topChunksLen = len(dataset)
			}
			prompt := "Extract comprehensive summary of all FINANCE related events.\n\n" + dataset[:topChunksLen]

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// 1. Embed dataset (cost of embeddings)
				embTokens := approximateTokenCount(dataset)
				embCost := (embTokens / 1000000.0) * PriceEmbeddingPer1M

				// 2. Pro model cost (only top 5 chunks)
				proTokens := approximateTokenCount(prompt)
				proCost := (proTokens / 1000000.0) * PriceProInputPer1M

				totalTokens := embTokens + proTokens
				totalCost := embCost + proCost

				b.ReportMetric(totalTokens, "tokens/op")
				b.ReportMetric(totalCost, "USD/op")
			}
		})
	}
}
