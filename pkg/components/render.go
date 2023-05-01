package components

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"text/template"

	templates "github.com/stolostron/backplane-operator/pkg/template"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// Renderer is a helm template renderer for a fs.FS.
type Renderer struct {
	componentName string
	dir           string
	fileSystem    fs.FS
	templates     []*template.Template
	ready         bool
}

// NewFileTemplateRenderer creates a TemplateRenderer with the given parameters and returns a pointer to it.
// helmChartDirPath must be an absolute file path to the root of the helm charts.
func NewGenericRenderer(fileSystem fs.FS, componentName string) *Renderer {
	return &Renderer{
		componentName: componentName,
		fileSystem:    fileSystem,
	}
}

// Init reads the files from the FS into templates
func (r *Renderer) Init() error {
	if err := r.loadTemplates(); err != nil {
		return err
	}

	r.ready = true
	return nil
}

func (r *Renderer) PrintTemplates(wr io.Writer, configuration Values) error {
	renderedStrings, err := r.renderTemplates(configuration)
	if err != nil {
		return err
	}
	for _, f := range renderedStrings {
		fmt.Fprintf(wr, "%s---\n", f)
	}
	return nil
}

// renderTemplates applies values to templates and returns resources in string form
func (r *Renderer) renderTemplates(configuration Values) ([]string, error) {
	if !r.ready {
		return nil, fmt.Errorf("renderer %s is not initialized", r.componentName)
	}

	rendered := []string{}
	for _, t := range r.templates {
		var buf bytes.Buffer
		err := t.Execute(&buf, configuration)
		if err != nil {
			return nil, fmt.Errorf("execute template: %v", err)
		}
		if buf.Len() == 0 {
			return nil, fmt.Errorf("rendered template is empty")
		}
		rendered = append(rendered, buf.String())
	}
	return rendered, nil
}

// Render applies values to templates and returns resources in unstructured form
func (r *Renderer) Render(values Values) ([]*unstructured.Unstructured, error) {
	var templates []*unstructured.Unstructured
	renderedStrings, err := r.renderTemplates(values)
	if err != nil {
		return nil, err
	}

	for _, s := range renderedStrings {
		unstructured := &unstructured.Unstructured{}
		if err = yaml.Unmarshal([]byte(s), unstructured); err != nil {
			return nil, fmt.Errorf("unmarshal to unstructured: %v", err)
		}
		// Add namespace to namespaced resources
		switch unstructured.GetKind() {
		case "Deployment", "ServiceAccount", "Role", "RoleBinding", "Service", "ConfigMap", "Route":
			unstructured.SetNamespace(values.Namespace)
		}
		templates = append(templates, unstructured)
	}

	return templates, nil
}

// loadTemplates reads all files from the filesystem and creates templates for each
func (r *Renderer) loadTemplates() error {
	// fnames, err := fs.Glob(r.fileSystem, path.Join(r.dir, "*.yaml"))
	fnames, err := getFilesRecursive(r.fileSystem, ".")
	if err != nil {
		return fmt.Errorf("list files: %v", err)
	}
	if len(fnames) == 0 {
		return errors.New("no files found")
	}

	var templates []*template.Template
	for _, f := range fnames {
		b, err := fs.ReadFile(r.fileSystem, f)
		if err != nil {
			return fmt.Errorf("read file: %v", err)
		}
		if len(b) == 0 {
			return fmt.Errorf("empty file: %s", f)
		}
		t := template.Must(template.New(f).Parse(string(b)))
		templates = append(templates, t)
	}
	r.templates = templates
	return nil
}

//go:embed templates
var templateFS embed.FS

func getFilesRecursive(f fs.FS, root string) ([]string, error) {
	res := []string{}
	err := fs.WalkDir(f, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		res = append(res, path)
		return nil
	})
	return res, err
}

func RenderCRDs() ([]*unstructured.Unstructured, error) {
	var crds []*unstructured.Unstructured
	files, err := loadStaticFiles(templates.CRDFS)
	if err != nil {
		return nil, err
	}
	for _, bfile := range files {
		crd := &unstructured.Unstructured{}
		if err = yaml.Unmarshal(bfile, crd); err != nil {
			return nil, fmt.Errorf("unmarshal to unstructured: %v", err)
		}
		crds = append(crds, crd)
	}
	return crds, nil
}

func PrintCRDs(wr io.Writer) error {
	files, err := loadStaticFiles(templates.CRDFS)
	if err != nil {
		return err
	}
	for _, f := range files {
		fmt.Fprintf(wr, "%s---\n", f)
	}
	return nil
}

func loadStaticFiles(fSys fs.FS) ([][]byte, error) {
	var templates [][]byte
	fnames, err := getFilesRecursive(fSys, ".")
	if err != nil {
		return nil, fmt.Errorf("list files: %v", err)
	}

	for _, f := range fnames {
		b, err := fs.ReadFile(fSys, f)
		if err != nil {
			return nil, fmt.Errorf("read file: %v", err)
		}

		templates = append(templates, b)
	}
	return templates, nil
}
