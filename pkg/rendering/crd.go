// Copyright Contributors to the Open Cluster Management project

package renderer

import (
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

//go:embed crds/*
var crdEmbed embed.FS

// RenderCRDs returns a list of unstructured CRD resources
func RenderCRDs() ([]*apiextensionsv1.CustomResourceDefinition, error) {
	var crds []*apiextensionsv1.CustomResourceDefinition

	err := fs.WalkDir(crdEmbed, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// crd := &unstructured.Unstructured{}
		crd := &apiextensionsv1.CustomResourceDefinition{}
		bytes, err := crdEmbed.ReadFile(path)
		if err != nil {
			return err
		}
		if err = yaml.Unmarshal(bytes, crd); err != nil {
			return err
		}
		if crd.Kind != "CustomResourceDefinition" || crd.Spec.Names.Kind == "" || crd.Spec.Group == "" {
			return fmt.Errorf("error reading file %s: not a valid CRD", d.Name())
		}
		crds = append(crds, crd)
		return nil
	})
	if err != nil {
		return crds, err
	}
	return crds, nil
}

func RenderCRDs(crdDir string) ([]*unstructured.Unstructured, []error) {
	var crds []*unstructured.Unstructured
	errs := []error{}

	if val, ok := os.LookupEnv("DIRECTORY_OVERRIDE"); ok {
		crdDir = path.Join(val, crdDir)
	}

	// Read CRD files
	err := filepath.Walk(crdDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		crd := &unstructured.Unstructured{}
		if info == nil || info.IsDir() {
			return nil
		}
		bytesFile, e := ioutil.ReadFile(path)
		if e != nil {
			errs = append(errs, fmt.Errorf("%s - error reading file: %v", info.Name(), err.Error()))
		}
		if err = yaml.Unmarshal(bytesFile, crd); err != nil {
			errs = append(errs, fmt.Errorf("%s - error unmarshalling file to unstructured: %v", info.Name(), err.Error()))
		}
		crds = append(crds, crd)
		return nil
	})
	if err != nil {
		return crds, errs
	}

	return crds, errs
}
