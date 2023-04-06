// Copyright Contributors to the Open Cluster Management project

package controllers

// import (
// 	"reflect"
// 	"testing"

// 	mcev1 "github.com/stolostron/backplane-operator/api/v1"
// 	"github.com/stolostron/backplane-operator/pkg/status"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// )

// func TestDiscoveryComponent_GetStatusReporter(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		mce  *mcev1.MultiClusterEngine
// 		want reflect.Type
// 	}{
// 		{
// 			name: "test",
// 			mce: &mcev1.MultiClusterEngine{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name: "mce",
// 				},
// 				Spec: mcev1.MultiClusterEngineSpec{
// 					Overrides: &mcev1.Overrides{
// 						Components: []mcev1.ComponentConfig{
// 							{Name: "discovery", Enabled: true},
// 						},
// 					},
// 				},
// 			},
// 			want: reflect.TypeOf(status.DeploymentStatus{}),
// 		},
// 		{
// 			name: "test",
// 			mce: &mcev1.MultiClusterEngine{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name: "mce",
// 				},
// 				Spec: mcev1.MultiClusterEngineSpec{
// 					Overrides: &mcev1.Overrides{
// 						Components: []mcev1.ComponentConfig{
// 							{Name: "discovery", Enabled: true},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			dc := &DiscoveryComponent{}
// 			if got := dc.GetStatusReporter(tt.mce); reflect.TypeOf(got) != tt.want {
// 				t.Errorf("type is %v", reflect.TypeOf(got))

// 			}
// 		})
// 	}
// }

// func Test_Component_Statuses(t *testing.T) {
// 	k8sClient := fake.NewClientBuilder().Build()
// 	tracker := &status.StatusTracker{Client: k8sClient}

// 	tracker.AddComponent()

// }
