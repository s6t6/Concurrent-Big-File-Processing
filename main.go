package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type data struct {
	key   string
	value string
}

func main() {
	start := time.Now()

	pwd, _ := os.Getwd()
	Unprocessed := pwd + "/Unprocessed-Passwords"
	Processed := pwd + "/Processed"
	Index := pwd + "/Index"
	files, err := os.Open(Unprocessed)

	if err != nil {
		fmt.Printf("there was an error during opening of the Unprocessed-Passwords directory : %v \n", err)
	}

	defer files.Close()

	fileInfos, err := files.ReadDir(-1)

	if err != nil {
		fmt.Printf("there was an error during access to the directory : %v \n", err)
	}

	wg := sync.WaitGroup{}
	channel := make(chan []data, 3)
	for k, entry := range fileInfos {
		ReadFile(k, entry, Index, Processed, Unprocessed, &wg, channel)
		err2 := os.Rename(Unprocessed+"/"+entry.Name(), Processed+"/"+entry.Name())

		if err2 != nil {
			fmt.Printf("there was an error deleting the file : %v \n", err2)
		}
	}

	go func() {
		wg.Wait()
		close(channel)
		fmt.Printf("Processing ended. Elapsed time: %v \n", time.Since(start)/time.Second)
	}()

	m := make(map[string][]string)

	for item := range channel {
		for _, elem := range item {

			//fmt.Printf("key: %v value: %v \n", elem.key, elem.value)
			m[elem.key] = append(m[elem.key], elem.value)
		}
	}

	var wg2 sync.WaitGroup

	for key, value := range m {
		wg2.Add(1)
		//fmt.Printf("key: %v value: %v \n", key, value)
		go func(key string, value []string) {
			WriteToIndex(key, value)
			wg2.Done()
		}(key, value)
	}

	wg2.Wait()
	fmt.Printf("Elapsed Time: %v", time.Since(start)/time.Second)
}

func WriteToIndex(key string, value []string) {

	var wg sync.WaitGroup

	if _, err := os.Stat(key); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(key, os.ModePerm)
		if err != nil {
			fmt.Printf("some error: %v\n", err)
		}
	}
	f, err1 := os.OpenFile(key+"/1.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	if err1 != nil {
		fmt.Printf("there was an error while creating or accessing the file : %v \n", err1)
	}
	defer f.Close()

	wg.Add(1)

	go func(value []string) {
		for _, val := range value {
			f.WriteString(val)
		}
		wg.Done()
	}(value)
	wg.Wait()
}

func ReadFile(route int, entry fs.DirEntry, Index string, Processed string, Unprocessed string, wg *sync.WaitGroup, channel chan<- []data) {

	file, err := os.Open(Unprocessed + "/" + entry.Name())
	//channel1 := make(chan data)
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
	channel chan<- []data) {
	defer wg.Done()
	var m data // struct with key and value strings. Holds the path of the target file and the line to be written
	var slice []data
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

		m.key = path
		m.value = line + " | " + hex.EncodeToString(md5hash[:]) + " | " + hex.EncodeToString(sha1hash[:]) +
			" | " + hex.EncodeToString(sha256hash[:]) + " | " + name + "\n"

		slice = append(slice, m)
	}
	fmt.Printf("goroutine %v's child %v finished \n", route, selfroute)
	channel <- slice
}

func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func ProcessFile(entry *fs.DirEntry, Index string, Processed string, i int, Unprocessed string, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Open(Unprocessed + "/" + (*entry).Name())

	if err != nil {
		fmt.Printf("there was an error during access to a file in the directory : %v \n", err)
	}

	defer file.Close()

	var mu sync.Mutex
	reader := bufio.NewReader(file)
	line, err := Readln(reader)
	a := 0
	for err == nil {

		alphanumeric := regexp.MustCompile("^[a-zA-Z0-9_]*$")
		isAlphanumeric := alphanumeric.MatchString(string(line[0]))
		path := Processed + "/symbol"
		md5hash := md5.Sum([]byte(line))
		sha1hash := sha1.Sum([]byte(line))
		sha256hash := sha256.Sum256([]byte(line))

		if isAlphanumeric {
			path = Processed + "/" + string(line[0])
		}

		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(path, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}

		mu.Lock()
		f, err1 := os.OpenFile(path+"/1.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		mu.Unlock()

		if err1 != nil {
			fmt.Printf("there was an error while creating or accessing the file : %v \n", err)
		}

		defer f.Close()

		f.WriteString(line + " | " + hex.EncodeToString(md5hash[:]) + " | " + hex.EncodeToString(sha1hash[:]) +
			" | " + hex.EncodeToString(sha256hash[:]) + " | " + (*entry).Name() + "\n")
		line, err = Readln(reader)
		fmt.Printf("line number: %v \n password: %v \n", a, line)
		a++

	}
	fmt.Printf("goroutine: %v \n file: %v \n ----Done----- \n", i, (*entry).Name())
	err2 := os.Remove(Unprocessed + "/" + (*entry).Name())

	if err2 != nil {
		fmt.Printf("there was an error during deletion of the file: %v \n", err2)
	}
}
