package scan

import (
	"time"
)

type StreamConfig struct {
	Enabled    bool
	IntervalMs int
}

type StreamEmitter func([]FileEntry) error

type StreamingProcessor struct {
	cfg       *StreamConfig
	batchSize int
	emitter   StreamEmitter
	lastEmit  time.Time
	interval  int
}

func NewStreamingProcessor(cfg *StreamConfig, emitter StreamEmitter) *StreamingProcessor {
	intervalMs := 250
	if cfg != nil && cfg.IntervalMs > 0 {
		intervalMs = cfg.IntervalMs
	}
	return &StreamingProcessor{
		cfg:       cfg,
		batchSize: 50,
		emitter:   emitter,
		lastEmit:  time.Now(),
		interval:  intervalMs,
	}
}

func (s *StreamingProcessor) ShouldEmit() bool {
	if s.cfg == nil || !s.cfg.Enabled {
		return false
	}
	elapsed := time.Since(s.lastEmit)
	return elapsed >= time.Duration(s.cfg.IntervalMs)*time.Millisecond
}

func (s *StreamingProcessor) EmitBatch(entries []FileEntry) error {
	if s.cfg == nil || !s.cfg.Enabled {
		return nil
	}
	if s.emitter == nil {
		return nil
	}
	if err := s.emitter(entries); err != nil {
		return err
	}
	s.lastEmit = time.Now()
	return nil
}

func (s *StreamingProcessor) IsEnabled() bool {
	return s.cfg != nil && s.cfg.Enabled
}
