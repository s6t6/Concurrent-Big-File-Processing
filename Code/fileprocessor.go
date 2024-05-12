package code

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"sync"
)

var ChunkSize int = 1024 * 128

func processChunk(paragraph string, Index string, name string, wg *sync.WaitGroup,
	channel chan<- []Data) {
	defer wg.Done()
	var m Data // struct with key and value strings. Holds the path of the target file and the line to be written
	var slice []Data
	lines := strings.Split(paragraph, "\n")

	for _, line := range lines {

		if len(line) == 0 {
			continue
		}
		alphanumeric := regexp.MustCompile("^[a-zA-Z0-9_]*$")
		upperCase := regexp.MustCompile("^[A-Z]*$")
		isAlphanumeric := alphanumeric.MatchString(string(line[0]))
		isUpperCase := upperCase.MatchString(string(line[0]))
		path := Index + "/symbol"
		md5hash := md5.Sum([]byte(line))
		sha1hash := sha1.Sum([]byte(line))
		sha256hash := sha256.Sum256([]byte(line))

		if isAlphanumeric {
			path = Index + "/" + string(line[0])
			if isUpperCase {
				path = path + "_"
			}
		}

		m.Key = path
		m.Value = line + " | " + hex.EncodeToString(md5hash[:]) +
			" | " + hex.EncodeToString(sha1hash[:]) +
			" | " + hex.EncodeToString(sha256hash[:]) +
			" | " + name + "\n"

		slice = append(slice, m)
	}

	channel <- slice
}

func processFile(
	entry fs.DirEntry,
	Index string,
	Unprocessed string,
	wg *sync.WaitGroup,
	channel chan<- []Data) {

	file, err := os.Open(Unprocessed + "/" + entry.Name())
	if err != nil {
		fmt.Printf("there was an error during access to a file in the directory : %v \n", err)
	}

	buf := make([]byte, ChunkSize)

	offset := 0
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	for i := 0; i <= int(fileSize)/(ChunkSize); i++ {
		wg.Add(1)
		n, err2 := file.ReadAt(buf, int64(offset))

		if n == 0 {
			if err2 != nil {
				fmt.Println(err)
				break
			}
			if err2 == io.EOF {
				break
			}
			return
		}

		for i := n - 1; i >= 0; i-- {

			if buf[i] == '\n' {
				n = i
				break
			}
		}
		val := string(buf[:n])

		go processChunk(val, Index, entry.Name(), wg, channel)
		offset += n + 1
	}
	file.Close()

}
