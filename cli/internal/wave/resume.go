package wave

// ComputeResume determines what action to take when resuming a spec.
// It scans all waves and picks the appropriate resume action based on state:
//   - If any wave is blocked, return "blocked" with that wave number.
//   - If any wave is in_progress, return "continue-wave" with that wave number.
//   - If all waves are complete, return "done".
//   - If the last wave is complete, return "next-wave" with next wave number.
//   - If no waves exist, return "next-wave" with wave 1.
func ComputeResume(specDir string) ResumeState {
	waves, err := ScanWaves(specDir)
	if err != nil || len(waves) == 0 {
		return ResumeState{
			Action:  "next-wave",
			WaveNum: 1,
			Reason:  "no waves found, start first wave",
		}
	}

	// Check for blocked waves first (highest priority)
	for _, w := range waves {
		if w.Status == "blocked" {
			return ResumeState{
				Action:  "blocked",
				WaveNum: w.WaveNum,
				Reason:  "wave is blocked by checkpoint failure",
			}
		}
	}

	// Check for in-progress waves
	for _, w := range waves {
		if w.Status == "in_progress" {
			return ResumeState{
				Action:  "continue-wave",
				WaveNum: w.WaveNum,
				Reason:  "wave has execution runs but no passing checkpoint",
			}
		}
	}

	// Check for pending waves
	for _, w := range waves {
		if w.Status == "pending" {
			return ResumeState{
				Action:  "continue-wave",
				WaveNum: w.WaveNum,
				Reason:  "wave exists but has no runs yet",
			}
		}
	}

	// All waves are complete
	lastWave := waves[len(waves)-1]
	return ResumeState{
		Action:  "done",
		WaveNum: lastWave.WaveNum,
		Reason:  "all waves complete",
	}
}
