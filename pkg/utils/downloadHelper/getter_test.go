package downloadHelper

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/goleak"
	"gotest.tools/v3/assert"
)

// zipFixture builds a one-file zip archive in memory.
func zipFixture(t *testing.T, name, content string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	assert.NilError(t, err)
	_, err = w.Write([]byte(content))
	assert.NilError(t, err)
	assert.NilError(t, zw.Close())

	return buf.Bytes()
}

// TestDownload exercises the download-and-unpack path against a local server.
// It used to fetch a GitHub archive over the network, which made the test fail
// on every run once that URL stopped resolving, and made it dependent on
// internet access in CI besides.
func TestDownload(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")) // https://github.com/census-instrumentation/opencensus-go/issues/1191

	archive := zipFixture(t, "hello.txt", "payload")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	dst := t.TempDir()

	err := Download(srv.URL+"/fixture.zip", dst)
	assert.NilError(t, err)

	got, err := os.ReadFile(filepath.Join(dst, "hello.txt"))
	assert.NilError(t, err)
	assert.Equal(t, "payload", string(got))
}

// TestDownloadUnreachableSource asserts the error is surfaced rather than
// swallowed, so a broken app-store URL fails loudly.
func TestDownloadUnreachableSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	err := Download(srv.URL+"/missing.zip", t.TempDir())
	assert.Assert(t, err != nil, "a 404 from the source must be reported as an error")
}
