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

// Hypothetical pricing rates (USD per 1 Million tokens) based on Gemini pricing guidelines.
const (
	PriceProInputPer1M   = 3.50  // gemini-2.5-pro approx input cost
	PriceFlashInputPer1M = 0.075 // gemini-2.5-flash approx input cost
	PriceEmbeddingPer1M  = 0.02  // text-embedding-004 approx cost
)

func getModelPrice(modelName string) float64 {
	switch modelName {
	case "gemini-2.5-pro":
		return PriceProInputPer1M
	case "gemini-2.5-flash", "gemini-2.5-flash-lite":
		return PriceFlashInputPer1M
	default:
		// Defaulting to Flash if unknown
		return PriceFlashInputPer1M
	}
}

var contextSizes = []int{10, 100, 1000, 10000, 100000, 1000000} // Multipliers for context generation (from small to very large)

func BenchmarkCost_StandardRouting_Pro(b *testing.B) {
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

func BenchmarkCost_StandardRouting_Flash(b *testing.B) {
	for _, size := range contextSizes {
		b.Run(fmt.Sprintf("ContextSize_%d", size), func(b *testing.B) {
			dataset := data.GenerateContext(size)
			// Standard routing sends everything to Flash
			prompt := "Extract comprehensive summary of all FINANCE related events.\n\n" + dataset

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tokens := approximateTokenCount(prompt)
				cost := (tokens / 1000000.0) * PriceFlashInputPer1M

				b.ReportMetric(tokens, "tokens/op")
				b.ReportMetric(cost, "USD/op")
			}
		})
	}
}

func BenchmarkCost_RAGRouting(b *testing.B) {
	orchestratorPrice := getModelPrice(ai.OrchestratorModelName)
	synthesisPrice := getModelPrice(ai.FinalSynthesisModelName)

	for _, size := range contextSizes {
		b.Run(fmt.Sprintf("ContextSize_%d", size), func(b *testing.B) {
			dataset := data.GenerateContext(size)

			// 1. Orchestration Cost (Fixed overhead per task)
			// Includes System Instructions + Prompt + Python Script Generation
			orchInputTokens := 650.0  // Approx tokens for complex RAG instructions
			orchOutputTokens := 500.0 // Approx tokens for generated Python script
			orchCost := ((orchInputTokens + orchOutputTokens) / 1000000.0) * orchestratorPrice

			// 2. Embedding Cost (Scales with dataset)
			embTokens := approximateTokenCount(dataset)
			embCost := (embTokens / 1000000.0) * PriceEmbeddingPer1M

			// 3. Final Synthesis Cost (Dynamic chunk extraction)
			// Since the script dynamically decides the number of chunks based on relevance,
			// we simulate an average case of ~6-7 highly relevant chunks (approx 2000 chars) 
			// passing the dynamic threshold.
			topChunksLen := 2000
			if len(dataset) < topChunksLen {
				topChunksLen = len(dataset)
			}
			prompt := "Extract comprehensive summary of all FINANCE related events.\n\n" + dataset[:topChunksLen]
			synthTokens := approximateTokenCount(prompt)
			synthCost := (synthTokens / 1000000.0) * synthesisPrice

			totalTokens := orchInputTokens + orchOutputTokens + embTokens + synthTokens
			totalCost := orchCost + embCost + synthCost

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.ReportMetric(totalTokens, "tokens/op")
				b.ReportMetric(totalCost, "USD/op")
			}
		})
	}
}
