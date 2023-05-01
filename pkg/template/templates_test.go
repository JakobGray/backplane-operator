package templates

import (
	"io/fs"
	"testing"

	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

func TestPrintDiscoveryFiles(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	_, err := fs.Stat(DiscoveryFS, "components/discovery/discovery-operator_v1_service.yaml")
	g.Expect(err).To(BeNil())
	// fmt.Println(info)

	fnames, err := fs.Glob(DiscoveryFS, "components/discovery/*.yaml")
	g.Expect(err).To(BeNil())
	g.Expect(fnames).To(Not(BeEmpty()))
	for _, f := range fnames {
		t.Log(f)
		b, err := fs.ReadFile(DiscoveryFS, f)
		g.Expect(err).To(BeNil())
		t.Log(string(b))
	}

	t.Fail()
}
