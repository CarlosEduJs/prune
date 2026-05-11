package scan

import (
	"testing"
	"time"
)

func TestStreamingProcessorShouldEmitUsesDefaultInterval(t *testing.T) {
	proc := NewStreamingProcessor(&StreamConfig{Enabled: true}, nil)
	if proc.ShouldEmit() {
		t.Fatalf("expected ShouldEmit to be false immediately")
	}

	proc.lastEmit = time.Now().Add(-251 * time.Millisecond)
	if !proc.ShouldEmit() {
		t.Fatalf("expected ShouldEmit to be true after default interval")
	}
}

func TestStreamingProcessorShouldEmitUsesConfiguredInterval(t *testing.T) {
	proc := NewStreamingProcessor(&StreamConfig{Enabled: true, IntervalMs: 10}, nil)
	proc.lastEmit = time.Now().Add(-5 * time.Millisecond)
	if proc.ShouldEmit() {
		t.Fatalf("expected ShouldEmit to be false before configured interval")
	}

	proc.lastEmit = time.Now().Add(-11 * time.Millisecond)
	if !proc.ShouldEmit() {
		t.Fatalf("expected ShouldEmit to be true after configured interval")
	}
}
