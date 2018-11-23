package process

import (
	"fmt"
	"github.com/pkg/xattr"
	"os"
	"strings"
)

func makeMeta(contentHash string, file os.FileInfo) string {
	return fmt.Sprintf("%s,%s,%s", contentHash, fileSizeStr(file), fileTimeStr(file))
}

func parseMeta(metaStr string) (string, string, string) {
	x := strings.Split(metaStr, ",")
	return x[0], x[1], x[2]
}

func getMetaSha256(metaStr string) (result string) {
	result, _, _ = parseMeta(metaStr)
	return
}

func updateMeta(path string, f os.FileInfo) string {
	sha256 := fileSha256(path)
	meta := makeMeta(sha256, f)
	if err := xattr.Set(path, "user.kuferek", []byte(meta)); err != nil {
		panic(err)
	}
	xattr.Set(path, "user.sha256", []byte(sha256))
	return meta
}

func readMeta(path string) (result string) {
	metab, _ := xattr.Get(path, "user.kuferek")
	if metab != nil {
		result = string(metab)
	}
	return
}

func validateMeta(metaStr string, file os.FileInfo, path string, content bool) (validMeta string) {
	if content {
		validMeta = updateMeta(path, file)
	} else {
		contentHash, _, _ := parseMeta(metaStr)
		validMeta = makeMeta(contentHash, file)
		if metaStr != validMeta {
			updateMeta(path, file)
		}
	}
	return
}
