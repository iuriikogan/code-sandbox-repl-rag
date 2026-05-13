package orchestrator

import (
	"fmt"
	"testing"

	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/data"
)

// approximateTokenCount estimates tokens by dividing character count by 4.
// This is a standard industry rule-of-thumb for LLM tokenization.
func approximateTokenCount(text string) float64 {
	return float64(len(text)) / 4.0
}

// Hypothetical pricing rates (USD per 1 Million tokens) for Gemini 3.1 Series.
const (
	Price31ProInputPer1M       = 3.50  // gemini-3.1-pro
	Price31FlashInputPer1M     = 0.075 // gemini-3.1-flash
	Price31FlashLiteInputPer1M = 0.01  // gemini-3.1-flash-lite (estimated)
	PriceEmbeddingPer1M        = 0.02  // text-embedding-004
)

func getModelPrice(modelName string) float64 {
	switch modelName {
	case "gemini-2.5-pro", "gemini-3.1-pro", "gemini-3.1-pro-preview":
		return Price31ProInputPer1M
	case "gemini-2.5-flash", "gemini-3.1-flash", "gemini-3.1-flash-preview":
		return Price31FlashInputPer1M
	case "gemini-2.5-flash-lite", "gemini-3.1-flash-lite", "gemini-3.1-flash-lite-preview":
		return Price31FlashLiteInputPer1M
	default:
		return Price31FlashInputPer1M
	}
}

var contextSizes = []int{10, 100, 1000, 10000, 100000} // Multipliers for context generation

func BenchmarkCost_DirectLongContext_Pro(b *testing.B) {
	for _, size := range contextSizes {
		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			dataset := data.GenerateContext(size)
			tokens := approximateTokenCount(dataset)
			cost := (tokens / 1000000.0) * Price31ProInputPer1M
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.ReportMetric(cost, "USD/op")
			}
		})
	}
}

func BenchmarkCost_StandardRouting_Flash(b *testing.B) {
	for _, size := range contextSizes {
		b.Run(fmt.Sprintf("ContextSize_%d", size), func(b *testing.B) {
			dataset := data.GenerateContext(size)
			// Standard routing sends everything to Flash
			prompt := "Extract comprehensive summary of all FINANCE related events.\n\n" + dataset

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tokens := approximateTokenCount(prompt)
				cost := (tokens / 1000000.0) * Price31FlashInputPer1M

				b.ReportMetric(tokens, "tokens/op")
				b.ReportMetric(cost, "USD/op")
			}
		})
	}
}

func BenchmarkCost_RAGRouting(b *testing.B) {
	orchPrice := getModelPrice("gemini-2.5-flash")
	synthPrice := getModelPrice(ai.FinalSynthesisModelName)

	for _, size := range contextSizes {
		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			dataset := data.GenerateContext(size)
			
			// 1. Embedding everything
			embTokens := approximateTokenCount(dataset)
			embCost := (embTokens / 1000000.0) * PriceEmbeddingPer1M

			// 2. Orchestration overhead
			orchCost := (1150.0 / 1000000.0) * orchPrice

			// 3. Final Synthesis (Assume 10 relevant chunks)
			synthCost := (2500.0 / 1000000.0) * synthPrice

			totalCost := embCost + orchCost + synthCost
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.ReportMetric(totalCost, "USD/op")
			}
		})
	}
}

func BenchmarkCost_TieredDiscoveryRAG(b *testing.B) {
	triagePrice := Price31FlashLiteInputPer1M
	orchPrice := Price31FlashInputPer1M
	synthPrice := Price31ProInputPer1M

	for _, size := range contextSizes {
		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			dataset := data.GenerateContext(size)
			tokens := approximateTokenCount(dataset)

			// 1. Triage Phase (Scan 100% with Flash-Lite)
			triageCost := (tokens / 1000000.0) * triagePrice

			// 2. Embedding Phase (Only 10% of data after triage)
			embCost := ((tokens * 0.1) / 1000000.0) * PriceEmbeddingPer1M

			// 3. Orchestration overhead
			orchCost := (1150.0 / 1000000.0) * orchPrice

			// 4. Final Synthesis
			synthCost := (2500.0 / 1000000.0) * synthPrice

			totalCost := triageCost + embCost + orchCost + synthCost
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.ReportMetric(totalCost, "USD/op")
			}
		})
	}
}

func BenchmarkCost_MapReduce_Flash(b *testing.B) {
	flashPrice := Price31FlashInputPer1M

	for _, size := range contextSizes {
		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			dataset := data.GenerateContext(size)
			tokens := approximateTokenCount(dataset)
			
			// Map-Reduce: Divide into chunks and summarize each, then combine.
			// Assume 8k token chunks.
			numChunks := tokens / 8000.0
			if numChunks < 1 { numChunks = 1 }
			
			// Each chunk summarization call
			mapCost := (tokens / 1000000.0) * flashPrice
			
			// Final reduce call (summarizing the summaries)
			reduceCost := ((numChunks * 500.0) / 1000000.0) * flashPrice

			totalCost := mapCost + reduceCost
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.ReportMetric(totalCost, "USD/op")
			}
		})
	}
}
