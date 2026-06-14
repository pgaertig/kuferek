package process

import (
	"os"
	"time"
	"strconv"
	"encoding/binary"
	"crypto/sha256"
	"fmt"
	"io"
	"encoding/hex"
)

func fileTimeStr(file os.FileInfo) string {
	return file.ModTime().UTC().Format(time.RFC3339)
}

// humanizeBytes formats a byte count using IEC binary units (B, KiB, MiB, ...).
func humanizeBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func fileSizeStr(file os.FileInfo) string {
	return strconv.FormatInt(file.Size(),10)
}

func fileSha256(filePath string) (result string) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return
	}

	binarySize := make([]byte, 8)
	binary.LittleEndian.PutUint64(binarySize, uint64(fileStat.Size()))

	hash := sha256.New()

	/*hash.Write(binarySize)

	// sum 10k beginning of file
	if _, err = io.CopyN(hash, file, 10000); err != nil {
		return
	}

	// jump to 10k before end of file
	file.Seek(-100000, io.SeekEnd)

	//sum 10k end of file
	if _, err = io.CopyN(hash, file, 10000); err != nil {
		return
	}*/

	if _, err = io.Copy(hash, file); err != nil {
		return
	}

	result = hex.EncodeToString(hash.Sum(nil))
	return
}
