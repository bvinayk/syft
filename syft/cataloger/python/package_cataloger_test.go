package python

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/anchore/stereoscope/pkg/file"

	"github.com/anchore/syft/syft/pkg"
	"github.com/go-test/deep"
)

type pythonTestResolverMock struct {
	metadataReader io.Reader
	recordReader   io.Reader
	topLevelReader io.Reader
	metadataRef    *file.Reference
	recordRef      *file.Reference
	topLevelRef    *file.Reference
	contents       map[file.Reference]string
}

func newTestResolver(metaPath, recordPath, topPath string) *pythonTestResolverMock {
	metadataReader, err := os.Open(metaPath)
	if err != nil {
		panic(fmt.Errorf("failed to open metadata: %+v", err))
	}

	var recordReader io.Reader
	if recordPath != "" {
		recordReader, err = os.Open(recordPath)
		if err != nil {
			panic(fmt.Errorf("failed to open record: %+v", err))
		}
	}

	var topLevelReader io.Reader
	if topPath != "" {
		topLevelReader, err = os.Open(topPath)
		if err != nil {
			panic(fmt.Errorf("failed to open top level: %+v", err))
		}
	}

	var recordRef *file.Reference
	if recordReader != nil {
		ref := file.NewFileReference("test-fixtures/dist-info/RECORD")
		recordRef = &ref
	}
	var topLevelRef *file.Reference
	if topLevelReader != nil {
		ref := file.NewFileReference("test-fixtures/dist-info/top_level.txt")
		topLevelRef = &ref
	}
	metadataRef := file.NewFileReference("test-fixtures/dist-info/METADATA")
	return &pythonTestResolverMock{
		recordReader:   recordReader,
		metadataReader: metadataReader,
		topLevelReader: topLevelReader,
		metadataRef:    &metadataRef,
		recordRef:      recordRef,
		topLevelRef:    topLevelRef,
		contents:       make(map[file.Reference]string),
	}
}

func (r *pythonTestResolverMock) FileContentsByRef(ref file.Reference) (string, error) {
	switch ref.Path {
	case r.topLevelRef.Path:
		b, err := ioutil.ReadAll(r.topLevelReader)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case r.metadataRef.Path:
		b, err := ioutil.ReadAll(r.metadataReader)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case r.recordRef.Path:
		b, err := ioutil.ReadAll(r.recordReader)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return "", fmt.Errorf("invalid value given")
}

func (r *pythonTestResolverMock) MultipleFileContentsByRef(_ ...file.Reference) (map[file.Reference]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *pythonTestResolverMock) FilesByPath(_ ...file.Path) ([]file.Reference, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *pythonTestResolverMock) FilesByGlob(_ ...string) ([]file.Reference, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *pythonTestResolverMock) RelativeFileByPath(_ file.Reference, path string) (*file.Reference, error) {
	switch {
	case strings.Contains(path, "RECORD"):
		return r.recordRef, nil
	case strings.Contains(path, "top_level.txt"):
		return r.topLevelRef, nil
	default:
		return nil, fmt.Errorf("invalid RelativeFileByPath value given: %q", path)
	}
}

func TestPythonPackageWheelCataloger(t *testing.T) {
	tests := []struct {
		MetadataFixture string
		RecordFixture   string
		TopLevelFixture string
		ExpectedPackage pkg.Package
	}{
		{
			MetadataFixture: "test-fixtures/egg-info/PKG-INFO",
			RecordFixture:   "test-fixtures/egg-info/RECORD",
			TopLevelFixture: "test-fixtures/egg-info/top_level.txt",
			ExpectedPackage: pkg.Package{
				Name:         "requests",
				Version:      "2.22.0",
				Type:         pkg.PythonPkg,
				Language:     pkg.Python,
				Licenses:     []string{"Apache 2.0"},
				FoundBy:      "python-package-cataloger",
				MetadataType: pkg.PythonPackageMetadataType,
				Metadata: pkg.PythonPackageMetadata{
					Name:                 "requests",
					Version:              "2.22.0",
					License:              "Apache 2.0",
					Platform:             "UNKNOWN",
					Author:               "Kenneth Reitz",
					AuthorEmail:          "me@kennethreitz.org",
					SitePackagesRootPath: "test-fixtures",
					Files: []pkg.PythonFileRecord{
						{Path: "requests-2.22.0.dist-info/INSTALLER", Digest: &pkg.Digest{"sha256", "zuuue4knoyJ-UwPPXg8fezS7VCrXJQrAP7zeNuwvFQg"}, Size: "4"},
						{Path: "requests/__init__.py", Digest: &pkg.Digest{"sha256", "PnKCgjcTq44LaAMzB-7--B2FdewRrE8F_vjZeaG9NhA"}, Size: "3921"},
						{Path: "requests/__pycache__/__version__.cpython-38.pyc"},
						{Path: "requests/__pycache__/utils.cpython-38.pyc"},
						{Path: "requests/__version__.py", Digest: &pkg.Digest{"sha256", "Bm-GFstQaFezsFlnmEMrJDe8JNROz9n2XXYtODdvjjc"}, Size: "436"},
						{Path: "requests/utils.py", Digest: &pkg.Digest{"sha256", "LtPJ1db6mJff2TJSJWKi7rBpzjPS3mSOrjC9zRhoD3A"}, Size: "30049"},
					},
					TopLevelPackages: []string{"requests"},
				},
			},
		},
		{
			MetadataFixture: "test-fixtures/dist-info/METADATA",
			RecordFixture:   "test-fixtures/dist-info/RECORD",
			TopLevelFixture: "test-fixtures/dist-info/top_level.txt",
			ExpectedPackage: pkg.Package{
				Name:         "Pygments",
				Version:      "2.6.1",
				Type:         pkg.PythonPkg,
				Language:     pkg.Python,
				Licenses:     []string{"BSD License"},
				FoundBy:      "python-package-cataloger",
				MetadataType: pkg.PythonPackageMetadataType,
				Metadata: pkg.PythonPackageMetadata{
					Name:                 "Pygments",
					Version:              "2.6.1",
					License:              "BSD License",
					Platform:             "any",
					Author:               "Georg Brandl",
					AuthorEmail:          "georg@python.org",
					SitePackagesRootPath: "test-fixtures",
					Files: []pkg.PythonFileRecord{
						{Path: "../../../bin/pygmentize", Digest: &pkg.Digest{"sha256", "dDhv_U2jiCpmFQwIRHpFRLAHUO4R1jIJPEvT_QYTFp8"}, Size: "220"},
						{Path: "Pygments-2.6.1.dist-info/AUTHORS", Digest: &pkg.Digest{"sha256", "PVpa2_Oku6BGuiUvutvuPnWGpzxqFy2I8-NIrqCvqUY"}, Size: "8449"},
						{Path: "Pygments-2.6.1.dist-info/RECORD"},
						{Path: "pygments/__pycache__/__init__.cpython-38.pyc"},
						{Path: "pygments/util.py", Digest: &pkg.Digest{"sha256", "586xXHiJGGZxqk5PMBu3vBhE68DLuAe5MBARWrSPGxA"}, Size: "10778"},
					},
					TopLevelPackages: []string{"pygments", "something_else"},
				},
			},
		},
		{
			// in cases where the metadata file is available and the record is not we should still record there is a package
			// additionally empty top_level.txt files should not result in an error
			MetadataFixture: "test-fixtures/partial.dist-info/METADATA",
			TopLevelFixture: "test-fixtures/partial.dist-info/top_level.txt",
			ExpectedPackage: pkg.Package{
				Name:         "Pygments",
				Version:      "2.6.1",
				Type:         pkg.PythonPkg,
				Language:     pkg.Python,
				Licenses:     []string{"BSD License"},
				FoundBy:      "python-package-cataloger",
				MetadataType: pkg.PythonPackageMetadataType,
				Metadata: pkg.PythonPackageMetadata{
					Name:                 "Pygments",
					Version:              "2.6.1",
					License:              "BSD License",
					Platform:             "any",
					Author:               "Georg Brandl",
					AuthorEmail:          "georg@python.org",
					SitePackagesRootPath: "test-fixtures",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.MetadataFixture, func(t *testing.T) {
			resolver := newTestResolver(test.MetadataFixture, test.RecordFixture, test.TopLevelFixture)

			// note that the source is the record ref created by the resolver mock... attach the expected values
			test.ExpectedPackage.Source = []file.Reference{*resolver.metadataRef}
			if resolver.recordRef != nil {
				test.ExpectedPackage.Source = append(test.ExpectedPackage.Source, *resolver.recordRef)
			}

			if resolver.topLevelRef != nil {
				test.ExpectedPackage.Source = append(test.ExpectedPackage.Source, *resolver.topLevelRef)
			}
			// end patching expected values with runtime data...

			pyPkgCataloger := NewPythonPackageCataloger()

			actual, err := pyPkgCataloger.catalogEggOrWheel(resolver, *resolver.metadataRef)
			if err != nil {
				t.Fatalf("failed to catalog python package: %+v", err)
			}

			for _, d := range deep.Equal(actual, &test.ExpectedPackage) {
				t.Errorf("diff: %+v", d)
			}
		})
	}

}
