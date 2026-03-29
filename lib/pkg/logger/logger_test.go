package logger

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name string
		opt  LoggerOption
	}{
		{
			name: "Production Mode",
			opt: LoggerOption{
				IsProd: true,
			},
		},
		{
			name: "Development Mode",
			opt: LoggerOption{
				IsProd: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLogger(tt.opt)
			if l == nil {
				t.Fatal("NewLogger returned nil")
			}
			if l.GetZap() == nil {
				t.Fatal("GetZap returned nil")
			}
			if l.opt.IsProd != tt.opt.IsProd {
				t.Errorf("expected IsProd to be %v, got %v", tt.opt.IsProd, l.opt.IsProd)
			}

			// Ensure logging methods don't panic
			l.Info("test info message")
			l.Debug("test debug message")
			l.Warn("test warn message")
			l.Error("test error message")
			// We can't easily test Fatal as it calls os.Exit
		})
	}
}
