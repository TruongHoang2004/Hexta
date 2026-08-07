package validator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	log "gitlab.com/ecommercehub1/shared/pkg/logger"
)

func TestRegisterValidations(t *testing.T) {
	v := NewValidator()
	logger := log.NewLogger(log.LoggerOption{IsProd: false})

	RegisterValidations(v, logger)

	tests := []struct {
		name      string
		fieldName string
		value     interface{}
		tag       string
		wantErr   bool
	}{
		{
			name:      "valid vat rate name 0%",
			fieldName: "VatRateName",
			value:     "0%",
			tag:       "valid-einvoice-vat-rate-name",
			wantErr:   false,
		},
		{
			name:      "invalid vat rate name",
			fieldName: "VatRateName",
			value:     "invalid",
			tag:       "valid-einvoice-vat-rate-name",
			wantErr:   true,
		},
		{
			name:      "valid today",
			fieldName: "Date",
			value:     time.Now(),
			tag:       "valid-today",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Var(tt.value, tt.tag)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
