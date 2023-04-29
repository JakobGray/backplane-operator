package components

import (
	"embed"
	"reflect"
	"testing"
)

//go:embed testdata
var testFS embed.FS

func TestPrintDiscoveryFiles(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "testrun"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PrintDiscoveryFiles()
		})
	}
}

func Test_getFilesRecursive(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		testName      string
		componentName string
		want          []string
		wantErr       bool
	}{
		{
			testName:      "Get testcomponent files",
			componentName: "testdata/testcomponent",
			want:          []string{"testdata/testcomponent/test-sa.yaml"},
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := getFilesRecursive(testFS, tt.componentName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFilesRecursive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFilesRecursive() = %v, want %v", got, tt.want)
			}
		})
	}
}
