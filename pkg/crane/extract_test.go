package crane

import (
	"archive/tar"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/spf13/afero"
)

func TestExtract(t *testing.T) {
	fs := afero.NewMemMapFs()
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	img, err := random.Image(1024, 5)
	if err != nil {
		t.Fatal(err)
	}

	name := "/some/file"
	content := []byte("sentinel")

	tw := tar.NewWriter(tmp)
	if err := tw.WriteHeader(&tar.Header{
		Size: int64(len(content)),
		Name: name,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	tw.Flush()
	tw.Close()

	img, err = Append(img, tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	if err := Extract(fs, img, ExtractArgs{"/": "/some/*"}); err != nil {
		t.Fatal(err)
	}

	absFileName := filepath.Join("/", name)
	if _, err := fs.Stat(absFileName); os.IsNotExist(err) {
		t.Fatal(err)
	}

	fc, err := afero.ReadFile(fs, absFileName)
	if err != nil {
		t.Fatal(err)
	}

	if string(fc) != string(content) {
		t.Fatal("file content error")
	}

}
