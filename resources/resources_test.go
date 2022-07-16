package resources

import (
	"bytes"
	"crypto/sha256"
	"io"
	"os"
	"testing"
)

func TestGetFile(t *testing.T) {
	vfs := New()

	computeHash := func(rdr io.Reader) []byte {
		hash := sha256.New()
		if _, err := io.Copy(hash, rdr); err != nil {
			t.Errorf("error calculating sha256: %s", err.Error())
		}
		return hash.Sum(nil)
	}

	tmpl, err := vfs.GetFile("/favicon.ico")
	if err != nil {
		t.Errorf("could not retrieve \"/favicon.ico\": %s", err.Error())
	}

	var vfsBuf bytes.Buffer
	if err := tmpl.Execute(&vfsBuf, tmpl); err != nil {
		t.Errorf("could not read data: %s", err.Error())
	}
	vfsSha := computeHash(&vfsBuf)

	fd, err := os.Open("root/favicon.ico")
	if err != nil {
		t.Errorf("could not open file: %s", err.Error())
	}
	realFSSha := computeHash(fd)

	if len(vfsSha) == 0 || len(realFSSha) == 0 {
		t.Error("empty SHA detected")
	}

	if !bytes.Equal(vfsSha, realFSSha) {
		t.Errorf(
			"\"/favicon.ico\" accessed by both the traditional filesystem and the VFS don't match. realfs = %x, vfs = %x",
			realFSSha, vfsSha)
	}
}
