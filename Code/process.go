package code

import (
	"fmt"
	"io/fs"
	"os"
	"sync"
	"time"
)

// Color escape codes for CLI
var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Magenta = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

func Process(pwd string) {
	Unprocessed := pwd + "/Unprocessed-Passwords"
	Processed := pwd + "/Processed"
	Index := pwd + "/Index"
	start := time.Now()

	os.RemoveAll(Index)
	os.Mkdir(Index, fs.ModePerm)

	//time measurement
	var sec, mili int
	t := time.Since(start)
	sec = int(t / time.Second)
	mili = int(t/time.Millisecond) - sec*1000
	fmt.Printf(Magenta+"Contents of the Index folder deleted."+Cyan+" It took %v seconds %v milliseconds \n"+Reset, sec, mili)

	moveTime := time.Now()
	processed, err := os.Open(Processed)
	if err != nil {
		fmt.Printf(Red+"there was an error during opening of the Processed-Passwords directory : %v \n"+Reset, err)
		return
	}
	processedInfo, err := processed.ReadDir(-1)

	if err != nil {
		fmt.Printf(Red+"there was an error during access to the directory : %v \n"+Reset, err)
		return
	}
	var wg sync.WaitGroup
	for _, entry := range processedInfo {
		wg.Add(1)
		go func(entry fs.DirEntry) {
			os.Rename(Processed+"/"+entry.Name(), Unprocessed+"/"+entry.Name())
			wg.Done()
		}(entry)
	}
	wg.Wait()

	processed.Close()

	//time measurement
	t = time.Since(moveTime)
	sec = int(t / time.Second)
	mili = int(t/time.Millisecond) - sec*1000
	fmt.Printf(Magenta+"Contents of the Processed folder moved to Unprocessed-Passwords."+Cyan+" It took %v seconds %v milliseconds \n"+Reset, sec, mili)

	processTime := time.Now()
	files, err := os.Open(Unprocessed)

	if err != nil {
		fmt.Printf(Red+"there was an error during opening of the Unprocessed-Passwords directory : %v \n"+Reset, err)
		return
	}

	defer files.Close()

	fileInfos, err := files.ReadDir(-1)

	if err != nil {
		fmt.Printf(Red+"there was an error during access to the directory : %v \n"+Reset, err)
		return
	}

	channel := make(chan []Data, 3)

	for _, entry := range fileInfos {
		processFile(entry, Index, Unprocessed, &wg, channel)
		err2 := os.Rename(Unprocessed+"/"+entry.Name(), Processed+"/"+entry.Name())

		if err2 != nil {
			fmt.Printf(Red+"there was an error deleting the file : %v \n"+Reset, err2)
		}
	}

	go func() {
		wg.Wait()
		close(channel)

		//time measurement
		t = time.Since(processTime)
		sec = int(t / time.Second)
		mili = int(t/time.Millisecond) - sec*1000
		fmt.Printf(Magenta+"Processing ended. Elapsed time:"+Cyan+" %v seconds %v milliseconds \n"+Reset, sec, mili)
	}()

	m := make(map[string][]string)

	for item := range channel {
		for _, elem := range item {
			m[elem.Key] = append(m[elem.Key], elem.Value)
		}
	}

	var wg2 sync.WaitGroup

	writeTime := time.Now()
	for key, value := range m {
		wg2.Add(1)
		go func(key string, value []string) {
			writeToIndex(key, value)
			wg2.Done()
		}(key, value)
	}

	wg2.Wait()

	//time measurement
	t = time.Since(writeTime)
	sec = int(t / time.Second)
	mili = int(t/time.Millisecond) - sec*1000
	fmt.Printf(Magenta+"Writing ended. Elapsed time:"+Cyan+" %v seconds %v milliseconds \n"+Reset, sec, mili)
}
