package data

import (
	"math/rand"
	"strings"
	"time"
)

type ScenarioEvent struct {
	Content string
	Index   int
}

func MedicalScenario() []ScenarioEvent {
	return []ScenarioEvent{
		{Content: "Patient A (Male, 28) admitted in 1985 with severe unexplained abdominal pain and photosensitivity.", Index: 5200},
		{Content: "Medical record 1992: Patient A reports recurrent episodes of dark urine during periods of fasting.", Index: 25000},
		{Content: "Patient B (Daughter of A) clinical note 2010: Patient presents with atypical peripheral neuropathy.", Index: 55000},
		{Content: "Patient C (Son of B) medication history 2025: Patient prescribed Sulfonamide antibiotics.", Index: 150222},
		{Content: "Patient C (Son of B) ER Admission 2025: Admitted with acute neurovisceral crisis and dark urine after Sulfonamides.", Index: 150900},
		{Content: "Medical Research: Sulfonamides are triggers for acute neurovisceral attacks in Acute Intermittent Porphyria (AIP).", Index: 200000},
	}
}

func EngineeringScenario() []ScenarioEvent {
	return []ScenarioEvent{
		{Content: "Service Alpha code snippet: Adds custom X-Trace-Legacy-ID header to downstream requests.", Index: 10452},
		{Content: "Service cron-beta log: Daily synchronization job started at 02:00:00 UTC to Service Omega /v1/legacy.", Index: 45901},
		{Content: "Envoy Release Notes: Known issue #442 - memory leak when non-standard X-Trace headers are present.", Index: 89112},
		{Content: "Service Omega metrics: OOM-kill triggered on 4 replicas at 02:48 UTC.", Index: 120555},
		{Content: "Service cron-beta status: Batch job failed due to connection reset from Service Omega at 02:48:12 UTC.", Index: 180000},
	}
}

func GenerateUltraMassiveContext() string {
	totalChunks := 1200000
	noises := []string{
		"System check: all services operational.",
		"Employee ID 4829 logged in at 09:00.",
		"Service Gamma heartbeat: OK.",
		"Database backup completed at 01:00.",
		"HR Policy update: vacation limit is 5 days.",
		"Network latency within parameters (2ms).",
		"Patient record updated: all vitals normal.",
		"Routine security scan: no vulnerabilities found.",
	}
	eventMap := make(map[int]string)
	for _, e := range MedicalScenario() { eventMap[e.Index] = e.Content }
	for _, e := range EngineeringScenario() { eventMap[e.Index] = e.Content }
	var sb strings.Builder
	sb.Grow(50 * 1024 * 1024)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < totalChunks; i++ {
		if content, ok := eventMap[i]; ok {
			sb.WriteString(content)
		} else {
			sb.WriteString(noises[r.Intn(len(noises))])
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
