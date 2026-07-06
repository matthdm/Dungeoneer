package combat

// SimResult holds aggregate metrics from one benchmarker scenario run.
type SimResult struct {
	ScenarioName    string
	Iterations      int
	AvgDPSTick      float64
	AvgClearTimeSec float64
	SurvivalRate    float64 // 0.0–1.0 fraction of iterations player survived
	AvgDamageTaken  float64
	KillsPerMinute  float64
	StreakAvg       float64
	SkillUsageRatio map[string]float64 // artifact ID → fraction of total actions
}
