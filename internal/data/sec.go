package data

import (
	"fmt"
	"strings"
)

// GenerateSECContext simulates a massive SEC 10-K filing.
func GenerateSECContext() string {
	var sb strings.Builder
	sb.WriteString("FORM 10-K - ANNUAL REPORT\n")
	sb.WriteString("COMPANY: ALPHA-OMEGA CORP\n")
	sb.WriteString("FISCAL YEAR ENDED: DECEMBER 31, 2024\n\n")

	sb.WriteString("ITEM 1. BUSINESS\n")
	sb.WriteString(strings.Repeat("Alpha-Omega is a global leader in AI-driven energy systems. ", 100))
	sb.WriteString("\n\n")

	sb.WriteString("ITEM 1A. RISK FACTORS\n")
	sb.WriteString("We face significant competition in the fusion energy market. ")
	sb.WriteString("Regulatory changes in the us-central1 region may impact our Code Sandbox deployments. ")
	sb.WriteString(strings.Repeat("Market volatility could affect our revenue. ", 200))
	sb.WriteString("\n\n")

	sb.WriteString("ITEM 7. MANAGEMENT'S DISCUSSION AND ANALYSIS (MD&A)\n")
	for i := 1; i <= 10; i++ {
		sb.WriteString(fmt.Sprintf("Q%d Financial Performance: ", i%4+1))
		sb.WriteString("Revenue increased by 15% due to high demand for GPU-based cooling systems. ")
		sb.WriteString(fmt.Sprintf("Operating expenses were $%dM. ", 40+i))
		sb.WriteString("We successfully completed the GKE Sandbox migration project.\n")
	}

	return sb.String()
}
