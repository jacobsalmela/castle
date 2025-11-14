package perfmon

import (
	"fmt"
	"sort"
	"time"
)

// SystemProfiler tracks execution time of individual systems
type SystemProfiler struct {
	timings      map[string]time.Duration
	calls        map[string]int
	startTimes   map[string]time.Time
	enabled      bool
	lastReport   time.Time
	reportBuffer []string
}

var GlobalProfiler = &SystemProfiler{
	timings:    make(map[string]time.Duration),
	calls:      make(map[string]int),
	startTimes: make(map[string]time.Time),
	enabled:    true,
	lastReport: time.Now(),
}

// Start begins timing a system
func (p *SystemProfiler) Start(systemName string) {
	if !p.enabled {
		return
	}
	p.startTimes[systemName] = time.Now()
}

// End stops timing a system and records duration
func (p *SystemProfiler) End(systemName string) {
	if !p.enabled {
		return
	}
	if start, ok := p.startTimes[systemName]; ok {
		elapsed := time.Since(start)
		p.timings[systemName] += elapsed
		p.calls[systemName]++
		delete(p.startTimes, systemName)
	}
}

// Report generates and returns a performance report
func (p *SystemProfiler) Report() string {
	if !p.enabled || len(p.timings) == 0 {
		return ""
	}

	// Calculate total time
	var total time.Duration
	for _, dur := range p.timings {
		total += dur
	}

	// Sort systems by time (descending)
	type systemTime struct {
		name     string
		duration time.Duration
		calls    int
	}
	var systems []systemTime
	for name, dur := range p.timings {
		systems = append(systems, systemTime{name, dur, p.calls[name]})
	}
	sort.Slice(systems, func(i, j int) bool {
		return systems[i].duration > systems[j].duration
	})

	// Build report (top 10 slowest)
	report := fmt.Sprintf("\n=== Performance Report (%.1fms total) ===\n", float64(total.Microseconds())/1000.0)
	count := 10
	if len(systems) < count {
		count = len(systems)
	}
	for i := 0; i < count; i++ {
		sys := systems[i]
		pct := float64(sys.duration) / float64(total) * 100
		avgMs := float64(sys.duration.Microseconds()) / float64(sys.calls) / 1000.0
		report += fmt.Sprintf("%-30s %6.2fms (%5.1f%%) [%d calls, %.3fms avg]\n",
			sys.name, float64(sys.duration.Microseconds())/1000.0, pct, sys.calls, avgMs)
	}

	return report
}

// Reset clears all timing data
func (p *SystemProfiler) Reset() {
	p.timings = make(map[string]time.Duration)
	p.calls = make(map[string]int)
	p.startTimes = make(map[string]time.Time)
}

// ShouldReport checks if it's time to print a report (every 5 seconds)
func (p *SystemProfiler) ShouldReport() bool {
	if time.Since(p.lastReport) > 5*time.Second {
		p.lastReport = time.Now()
		return true
	}
	return false
}

// Enable turns profiling on
func (p *SystemProfiler) Enable() {
	p.enabled = true
}

// Disable turns profiling off
func (p *SystemProfiler) Disable() {
	p.enabled = false
}

// GetTopSlowest returns the N slowest systems
func (p *SystemProfiler) GetTopSlowest(n int) []string {
	type systemTime struct {
		name     string
		duration time.Duration
	}
	var systems []systemTime
	for name, dur := range p.timings {
		systems = append(systems, systemTime{name, dur})
	}
	sort.Slice(systems, func(i, j int) bool {
		return systems[i].duration > systems[j].duration
	})

	var result []string
	count := n
	if len(systems) < count {
		count = len(systems)
	}
	for i := 0; i < count; i++ {
		result = append(result, fmt.Sprintf("%s: %.2fms",
			systems[i].name,
			float64(systems[i].duration.Microseconds())/1000.0))
	}
	return result
}
