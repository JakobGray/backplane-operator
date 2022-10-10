// Copyright Contributors to the Open Cluster Management project

package renderer

import (
	"testing"
)

func TestRenderCRDs(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "Render CRDs directory",
			wantErr: nil,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RenderCRDs()
			if err != tt.wantErr {
				t.Errorf("RenderCRDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
