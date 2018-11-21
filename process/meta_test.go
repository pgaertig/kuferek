package process

import (
	"testing"
	"os"
	"strings"
)

func TestMakeMeta(t * testing.T) {
	fileInfo, err := os.Stat("../testdata/file1.txt")
	if err != nil {
		t.Error(err)
		return
	}

	meta := updateMeta("../testdata/file1.txt", fileInfo )
	if !strings.HasPrefix(meta,"11ff46be634069db0d303a12357d86480b42a60696b81f7bafeced8b01ee50cb,8,") {
		t.Error("Failed calculating meta: ", meta)
		return
	}

	println(meta)
}