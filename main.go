package main

import (
	code "1306170097/fileorg/project/Code"
	"fmt"
	"os"
	"sync"
	"time"
)

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
	channel := make(chan []code.Data, 3)

	for k, entry := range fileInfos {
		code.ReadFile(k, entry, Index, Processed, Unprocessed, &wg, channel)
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
			m[elem.Key] = append(m[elem.Key], elem.Value)
		}
	}

	var wg2 sync.WaitGroup

	for key, value := range m {
		wg2.Add(1)
		//fmt.Printf("key: %v value: %v \n", key, value)
		go func(key string, value []string) {
			code.WriteToIndex(key, value)
			wg2.Done()
		}(key, value)
	}

	wg2.Wait()
	fmt.Printf("Elapsed Time: %v", time.Since(start)/time.Second)
}
