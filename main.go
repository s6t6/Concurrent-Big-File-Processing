package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"regexp"
	"sync"
)

type data struct {
	key   string
	value string
}

func main() {

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
	writerMap := make(chan data)
	for i, entry := range fileInfos {
		wg.Add(1)
		fmt.Printf("goroutine: %v \n file: %v \n ----Started----- \n", i, entry.Name())
		go ProcessFileWithChan(entry, Index, Processed, i, Unprocessed, &wg, writerMap)

	}

	go func() {
		wg.Wait()
		close(writerMap)
	}()

	m := make(map[string]map[string]bool)

	for item := range writerMap {
		fmt.Printf("\n\nchannel output \nkey: %v \nvalue: %v \n", item.key, item.value)
		m[item.key][item.value] = true
	}

	var wg2 sync.WaitGroup

	for key, value := range m {
		f, err1 := os.OpenFile(key+"/1.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

		if err1 != nil {
			fmt.Printf("there was an error while creating or accessing the file : %v \n", err)
		}
		defer f.Close()

		wg2.Add(1)

		go func(value map[string]bool) {
			for val := range value {
				f.WriteString(val)
			}
			wg2.Done()
		}(value)
	}

	wg2.Wait()
}

func ProcessFileWithChan(entry fs.DirEntry, Index string, Processed string, i int, Unprocessed string, wg *sync.WaitGroup,
	writerMap chan<- data) {
	defer wg.Done()
	file, err := os.Open(Unprocessed + "/" + entry.Name())

	if err != nil {
		fmt.Printf("there was an error during access to a file in the directory : %v \n", err)
	}

	var m data
	reader := bufio.NewReader(file)
	line, err := Readln(reader)
	a := 0
	for err == nil {

		alphanumeric := regexp.MustCompile("^[a-zA-Z0-9_]*$")
		isAlphanumeric := alphanumeric.MatchString(string(line[0]))
		path := Index + "/symbol"
		md5hash := md5.Sum([]byte(line))
		sha1hash := sha1.Sum([]byte(line))
		sha256hash := sha256.Sum256([]byte(line))

		if isAlphanumeric {
			path = Index + "/" + string(line[0])
		}

		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(path, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}

		m.key = path
		m.value = line + " | " + hex.EncodeToString(md5hash[:]) + " | " + hex.EncodeToString(sha1hash[:]) +
			" | " + hex.EncodeToString(sha256hash[:]) + " | " + entry.Name() + "\n"
		fmt.Printf("key: %v value: %v \n", m.key, m.value)
		writerMap <- m

		line, err = Readln(reader)
		//fmt.Printf("line number: %v \n password: %v \n", a, line)
		a++

	}

	fmt.Printf("goroutine: %v \n file: %v \n ----Done----- \n", i, entry.Name())

	file.Close()
	err2 := os.Rename(Unprocessed+"/"+entry.Name(), Processed+"/"+entry.Name())

	if err2 != nil {
		fmt.Printf("there was an error deleting the file : %v \n", err2)
	}
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
