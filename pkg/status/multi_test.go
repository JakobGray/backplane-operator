// Copyright Contributors to the Open Cluster Management project
package status

import (
	"testing"

	"github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// type buggyClient struct {
// 	client.Client
// 	shouldError bool
// }

// func (cl buggyClient) Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
// 	if cl.shouldError {
// 		return apierrors.NewInternalError(fmt.Errorf("oops"))
// 	}
// 	return cl.Client.Get(ctx, key, obj)
// }

func TestNewMultiStatus(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// testNS := types.NamespacedName{Name: "test", Namespace: ""}

	sr := NewMultiStatus(
		types.NamespacedName{Name: "testComponent", Namespace: ""},
		[]StatusReporter{
			NewPresentStatus(types.NamespacedName{Name: "test"}, schema.GroupVersionKind{Group: "", Kind: "Namespace", Version: "v1"}),
			NewPresentStatus(types.NamespacedName{Name: "test-deployment", Namespace: "test"}, schema.GroupVersionKind{Group: "apps", Kind: "Deployment", Version: "v1"}),
		})

	sr2 := NewMultiStatus(
		types.NamespacedName{Name: "testComponent", Namespace: ""},
		[]StatusReporter{
			NewPresentStatus(types.NamespacedName{Name: "test"}, schema.GroupVersionKind{Group: "", Kind: "Namespace", Version: "v1"}),
			NewPresentStatus(types.NamespacedName{Name: "test-deployment", Namespace: "test"}, schema.GroupVersionKind{Group: "apps", Kind: "Deployment", Version: "v1"}),
			NewPresentStatus(types.NamespacedName{Name: "test-deployment2", Namespace: "test"}, schema.GroupVersionKind{Group: "apps", Kind: "Deployment", Version: "v1"}),
		})

	g.Expect(sr.GetName()).To(gomega.Equal("testComponent"))
	g.Expect(sr.GetKind()).To(gomega.Equal("Component"))
	g.Expect(sr.GetNamespace()).To(gomega.Equal(""))

	cl := fake.NewClientBuilder().WithObjects(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test"}},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "test-deployment", Namespace: "test"}},
	).Build()

	condition := sr.Status(cl)
	g.Expect(condition.Available).To(gomega.BeTrue(), "Status should be good because deployment exists")
	g.Expect(condition.Message).To(gomega.BeEmpty(), "Message is omitted when all components are available")

	condition2 := sr2.Status(cl)
	g.Expect(condition2.Available).To(gomega.BeFalse(), "Condition should not be available because of missing deployment")

	// cl = fake.NewClientBuilder().WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test"}}).Build()
	// condition = sr.Status(cl)
	// g.Expect(condition.Available).To(gomega.BeFalse(), "Status should not be good because namespace exists")

	// bc := buggyClient{cl, true}
	// condition = sr.Status(bc)
	// g.Expect(condition.Type).To(gomega.Equal("Unknown"), "Status should be unknown due to request error")

}
