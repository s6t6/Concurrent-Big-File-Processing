package code

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
)

func writeToIndex(key string, value []string) {

	var wg sync.WaitGroup

	if _, err := os.Stat(key); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(key, os.ModePerm)
		if err != nil {
			fmt.Printf(Yellow+"warning, file already exists: %v\n"+Reset, err)
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
		slices.Sort(values)
		wg.Done()
	}(value)
	wg.Wait()

	wg.Add(1)

	go func(values []string) {
		fileNumber := 1
		idx := 0
		lineNumber := 0
		var keyrange []string

		var metadata *os.File

		for {
			f, err1 := os.OpenFile(key+"/"+fmt.Sprint(fileNumber)+".txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

			if err1 != nil {
				fmt.Printf(Red+"there was an error while creating or accessing the file : %v \n"+Reset, err1)
			}
			for _, val := range values[idx:] {
				f.WriteString(val)
				idx++
				if lineNumber == 0 {
					keyrange = append(keyrange, fmt.Sprintf("%d,%s", fileNumber, strings.Split(val, " ")[0]))
				} else if lineNumber == 9999 {
					lineNumber = 0
					keyrange[len(keyrange)-1] = keyrange[len(keyrange)-1] + "," + strings.Split(val, " ")[0]
					break
				} else if idx+1 == len(values) {
					keyrange[len(keyrange)-1] = keyrange[len(keyrange)-1] + "," + strings.Split(val, " ")[0]
				}
				lineNumber++
			}
			f.Close()

			if idx == len(values) {
				break
			}
			fileNumber++
		}

		os.Remove(key + "/metadata.txt")
		metadata, _ = os.Create(key + "/metadata.txt")

		slices.Sort(keyrange)
		for _, k := range keyrange {
			metadata.WriteString(k + "\n")
		}
		metadata.Close()
		wg.Done()
	}(values)
	wg.Wait()
}
