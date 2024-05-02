package code

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"sync"
)

func WriteToIndex(key string, value []string) {

	var wg sync.WaitGroup

	if _, err := os.Stat(key); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(key, os.ModePerm)
		if err != nil {
			fmt.Printf("some error: %v\n", err)
		}
	}

	set := make(map[string]string)
	values := make([]string, len(set))
	wg.Add(1)

	go func(value []string) {
		for _, val := range value {
			s := strings.Split(val, " ")[0]
			set[s] = val
		}
		for _, value := range set {

			values = append(values, value)

		}
		wg.Done()
	}(value)
	wg.Wait()

	wg.Add(1)

	go func(values []string) {
		finished := false
		fileNumber := 1
		idx := 0
		for !finished {
			f, err1 := os.OpenFile(key+"/"+fmt.Sprint(fileNumber)+".txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

			if err1 != nil {
				fmt.Printf("there was an error while creating or accessing the file : %v \n", err1)
			}

			for i, val := range values[idx:] {
				f.WriteString(val)
				idx++
				if i == 9999 {
					break
				}
			}
			f.Close()
			if idx == len(values) {
				finished = true
				break
			}
			fileNumber++
		}
		wg.Done()
	}(values)
	wg.Wait()
}

func ReadFile(route int, entry fs.DirEntry, Index string, Processed string, Unprocessed string, wg *sync.WaitGroup, channel chan<- []Data) {

	file, err := os.Open(Unprocessed + "/" + entry.Name())
	//channel1 := make(chan Data)
	if err != nil {
		fmt.Printf("there was an error during access to a file in the directory : %v \n", err)
	}

	buf := make([]byte, 1024*16)

	offset := 0
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	for i := 0; i <= int(fileSize)/(1024*16); i++ {
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

		go ProcessFileWithChan(i, route, val, Index, Processed, Unprocessed, entry.Name(), wg, channel)
		offset += n + 1
	}
	fmt.Printf("goroutine %v finished \n", route)
	file.Close()

}

func print(chunks []byte) {

	paragraph := string(chunks)
	lines := strings.Split(paragraph, "\n")
	file, _ := os.OpenFile("kayit.txt", os.O_CREATE|os.O_APPEND, 0644)

	for i, line := range lines {
		file.WriteString(fmt.Sprintf("%vth line: %v \n", i, line))
	}
}

func ProcessFileWithChan(selfroute int, route int, paragraph string, Index string, Processed string, Unprocessed string, name string, wg *sync.WaitGroup,
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
		isAlphanumeric := alphanumeric.MatchString(string(line[0]))
		path := Index + "/symbol"
		md5hash := md5.Sum([]byte(line))
		sha1hash := sha1.Sum([]byte(line))
		sha256hash := sha256.Sum256([]byte(line))

		if isAlphanumeric {
			path = Index + "/" + string(line[0])
		}

		m.Key = path
		m.Value = line + " | " + hex.EncodeToString(md5hash[:]) + " | " + hex.EncodeToString(sha1hash[:]) +
			" | " + hex.EncodeToString(sha256hash[:]) + " | " + name + "\n"

		slice = append(slice, m)
	}
	fmt.Printf("goroutine %v's child %v finished \n", route, selfroute)
	channel <- slice
}
