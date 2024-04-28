package tt

import (
	"os"
	"testing"
)

func TestSingleDownload(t *testing.T) {
	if err := os.MkdirAll("testdata", 0755); err != nil {
		t.Fail()
	}
	defer os.RemoveAll("testdata")

	// this vid might die and should be replaced with another one
	if _, err := Download("6895025692002487557", &DownloadOpt{Directory: "testdata"}); err != nil {
		t.Fail()
	}
}
