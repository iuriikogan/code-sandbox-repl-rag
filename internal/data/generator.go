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
                {Content: "Medical record 1992: Patient A reports recurrent episodes of dark urine during periods of fasting.", Index: 15000},
                {Content: "Patient B (Daughter of A) clinical note 2010: Patient presents with atypical peripheral neuropathy.", Index: 35000},
                {Content: "Patient C (Son of B) medication history 2025: Patient prescribed Sulfonamide antibiotics.", Index: 50222},
                {Content: "Patient C (Son of B) ER Admission 2025: Admitted with acute neurovisceral crisis and dark urine after Sulfonamides.", Index: 65900},
                {Content: "Medical Research: Sulfonamides are triggers for acute neurovisceral attacks in Acute Intermittent Porphyria (AIP).", Index: 75000},
        }
}

func EngineeringScenario() []ScenarioEvent {
        return []ScenarioEvent{
                {Content: "Compliance Team update: Rule 44B mandates legacy trace tracking.", Index: 100},
                {Content: "Service Alpha code snippet: Adds custom X-Trace-Legacy-ID header to downstream requests to satisfy Rule 44B.", Index: 10452},
                {Content: "LaunchDarkly Audit: SRE Team Lead (Alice) enabled feature flag FF_ARCHIVE_SYNC at 01:00 UTC.", Index: 18000},
                {Content: "Service cron-beta log: Feature flag FF_ARCHIVE_SYNC is true. Including Archive data in sync.", Index: 20000},
                {Content: "Service cron-beta log: Payload size calculated at 5.2MB for current batch.", Index: 21500},
                {Content: "Service cron-beta log: Daily synchronization job started at 02:00:00 UTC to Service Omega /v1/legacy.", Index: 25901},
                {Content: "Envoy Release Notes: Known issue #442 - memory leak when non-standard X-Trace headers are present and payload exceeds 2MB.", Index: 49112},
                {Content: "Kubernetes Events: cgroup memory controller invoked OOM killer for pod service-omega-77x9.", Index: 58000},
                {Content: "Service Omega metrics: istio-proxy sidecar container terminated with reason OOMKilled at 02:48 UTC.", Index: 60555},
                {Content: "Service Omega metrics: OOM-kill triggered on 4 replicas at 02:48 UTC.", Index: 60556},
                {Content: "Service cron-beta status: Batch job failed due to connection reset from Service Omega at 02:48:12 UTC.", Index: 78000},
        }
}
func GenerateUltraMassiveContext(totalChunks int) string {
        noises := []string{
                "System check: Service Omega operational, memory stable.",
                "Employee ID 4829 logged in at 09:00.",
                "Service Gamma heartbeat: OK, no OOM-kill detected.",
                "Database backup completed at 01:00 for cron-beta.",
                "HR Policy update: vacation limit is 5 days.",
                "Network latency within parameters (2ms) for istio-proxy.",
                "Patient record updated: all vitals normal.",
                "Routine security scan: no vulnerabilities found in Envoy.",
                "Service Alpha metrics: payload size 1.5MB, within limits.",
                "LaunchDarkly Audit: Feature flag FF_NEW_UI disabled.",
                "Kubernetes Events: pod service-omega-11b2 started successfully.",
                "Compliance Team update: Rule 44A reviewed.",
        }
        eventMap := make(map[int]string)
        for _, e := range MedicalScenario() { eventMap[e.Index] = e.Content }
        for _, e := range EngineeringScenario() { eventMap[e.Index] = e.Content }
        var sb strings.Builder
        // Estimate size: ~80 bytes per line
        sb.Grow(totalChunks * 80)
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
