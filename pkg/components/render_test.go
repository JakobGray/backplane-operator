package components

import (
	"bytes"
	"embed"
	"io/fs"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"

	templates "github.com/stolostron/backplane-operator/pkg/template"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"

	"github.com/stolostron/backplane-operator/pkg/utils"
)

//go:embed testdata/testcomponent
var testFS embed.FS

//go:embed testdata/crds
var testCRDs embed.FS

func TestRenderFiles(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// NewGenericRenderer(testFS, "testcomponent")

	testImages := map[string]string{}
	for _, v := range utils.GetTestImages() {
		testImages[v] = "quay.io/test/test:latest"
	}

	testVals := Values{
		Namespace:  "multicluster-engine",
		Org:        "open-cluster-management",
		PullPolicy: "IfNotPresent",
		Images:     testImages,
		ProxyConfigs: map[string]string{
			"HTTP_PROXY":  "test1",
			"HTTPS_PROXY": "test2",
			"NO_PROXY":    "test3",
		},
		PullSecret: "testpullsecret",
		NodeSelector: map[string]string{
			"select": "test",
		},
		Tolerations: []corev1.Toleration{
			{
				Key:      "dedicated",
				Operator: "Exists",
				Effect:   "NoSchedule",
				Value:    "test",
			},
		},
	}

	resources, err := Discovery.Render(testVals)
	g.Expect(err).To(BeNil())
	g.Expect(len(resources)).To(BeNumerically(">", 0))
	for _, r := range resources {
		t.Logf("%s/%s", r.GetKind(), r.GetName())
		switch r.GetKind() {
		case "Deployment":
			deployment := &appsv1.Deployment{}
			g.Expect(runtime.DefaultUnstructuredConverter.FromUnstructured(r.Object, deployment)).To(Succeed())
			g.Expect(deployment.Spec.Template.Spec.NodeSelector).To(Equal(testVals.NodeSelector))
			g.Expect(deployment.Spec.Template.Spec.Tolerations).To(Equal(testVals.Tolerations))
			g.Expect(deployment.GetNamespace()).To(Equal(testVals.Namespace))
			for _, c := range deployment.Spec.Template.Spec.Containers {
				g.Expect(c.Image).To(Equal("quay.io/test/test:latest"))
				g.Expect(string(c.ImagePullPolicy)).To(Equal(testVals.PullPolicy))
				for _, proxyVar := range c.Env {
					switch proxyVar.Name {
					case "HTTP_PROXY":
						g.Expect(proxyVar.Value).To(Equal(testVals.ProxyConfigs["HTTP_PROXY"]))
					case "HTTPS_PROXY":
						g.Expect(proxyVar.Value).To(Equal(testVals.ProxyConfigs["HTTPS_PROXY"]))
					case "NO_PROXY":
						g.Expect(proxyVar.Value).To(Equal(testVals.ProxyConfigs["NO_PROXY"]))
					}
				}
			}

		case "AddOnDeploymentConfig":
			addonDep := &addonv1alpha1.AddOnDeploymentConfig{}
			g.Expect(runtime.DefaultUnstructuredConverter.FromUnstructured(r.Object, addonDep)).To(Succeed())
			g.Expect(addonDep.Spec.NodePlacement.NodeSelector).To(Equal(testVals.NodeSelector))
			g.Expect(addonDep.Spec.NodePlacement.Tolerations).To(Equal(testVals.Tolerations))

		default:
			// g.Expect(r.GetLabels())
		}

	}
	// t.Fail()
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

func TestRenderer_loadTemplates(t *testing.T) {
	type fields struct {
		componentName string
		fileSystem    fs.FS
		dir           string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Read files",
			fields: fields{
				componentName: "discovery",
				fileSystem:    templates.DiscoveryFS,
				dir:           "components/discovery",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Renderer{
				componentName: tt.fields.componentName,
				fileSystem:    tt.fields.fileSystem,
				dir:           tt.fields.dir,
			}
			if err := r.loadTemplates(); (err != nil) != tt.wantErr {
				t.Errorf("Renderer.loadTemplates() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(r.templates) == 0 {
				t.Errorf("No templates loaded")
			}
		})
	}
}

func TestPrintCRDs(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Print out CRDs",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wr := &bytes.Buffer{}
			if err := PrintCRDs(wr); (err != nil) != tt.wantErr {
				t.Errorf("PrintCRDs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if wr.Len() == 0 {
				t.Errorf("PrintCRDs() returned empty string")
			}
		})
	}
}

func TestPrintTemplates(t *testing.T) {
	tests := []struct {
		name      string
		wantErr   bool
		component *Renderer
		config    Values
	}{
		{
			name:      "Print out CRDs",
			component: NewGenericRenderer(testFS, "testcomponent"),
			wantErr:   false,
			config: Values{
				Namespace:  "multicluster-engine",
				Org:        "open-cluster-management",
				PullPolicy: "Always",
				Images: map[string]string{
					"discovery_operator": "quay.io/jakobgray/discovery-operator:latest",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.component.Init(); err != nil {
				t.Error("Couldn't initialize")
			}
			wr := &bytes.Buffer{}
			if err := tt.component.PrintTemplates(wr, tt.config); (err != nil) != tt.wantErr {
				t.Errorf("PrintCRDs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if wr.Len() == 0 {
				t.Errorf("PrintCRDs() returned empty string")
			}
			// t.Log(wr.String())
			// t.Fail()
		})
	}
}

func Test_loadStaticFiles(t *testing.T) {
	tests := []struct {
		name    string
		fSys    fs.FS
		wantLen int
		wantErr bool
	}{
		{
			name:    "Load CRDs",
			fSys:    testCRDs,
			wantLen: 2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadStaticFiles(tt.fSys)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadStaticFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("loadStaticFiles() = %v, want %v", got, tt.wantLen)
			}
		})
	}
}
