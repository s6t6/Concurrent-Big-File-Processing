package code

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func Search(value string, pwd string) {

	indexPath := pwd + "/Index/"
	firstChar := string(value[0])
	alphanumeric := regexp.MustCompile("^[a-zA-Z0-9_]*$")
	upperCase := regexp.MustCompile("^[A-Z]*$")
	isAlphanumeric := alphanumeric.MatchString(firstChar)
	isUpperCase := upperCase.MatchString(firstChar)
	var folder string

	if isAlphanumeric {
		if isUpperCase {
			folder = indexPath + firstChar + "_"
		} else {
			folder = indexPath + firstChar
		}
	} else {
		folder = indexPath + "symbol"
	}

	metadataPath := folder + "/metadata.txt"

	metadata, err := os.OpenFile(metadataPath, os.O_RDONLY, 0644)

	if err != nil {
		fmt.Printf(Red+"couldn't access the metadata file: %v \n"+Reset, err)
	}
	defer metadata.Close()

	var keyrange []metadataEntry

	scanner := bufio.NewScanner(metadata)

	for scanner.Scan() {
		s := scanner.Text()
		if s != "" {
			fileno, _ := strconv.Atoi(strings.Split(s, ",")[0])
			start := strings.Split(s, ",")[1]
			end := strings.Split(s, ",")[2]
			keyrange = append(keyrange, metadataEntry{file: fileno, start: start, end: end})
		}
	}
	var target string
	for _, k := range keyrange {
		if value >= k.start && value <= k.end {
			target = folder + "/" + strconv.Itoa(k.file) + ".txt"
			break
		}
	}
	if target == "" {
		fmt.Printf(Red+"'%v' doesn't exist in the files\n"+Reset, value)
		return
	}

	file, err := os.OpenFile(target, os.O_RDONLY, 0644)

	if err != nil {
		fmt.Printf(Red+"couldn't access the file: %v \n"+Reset, err)
		return
	}
	defer file.Close()

	scanner = bufio.NewScanner(file)
	var list []Data

	for scanner.Scan() {
		s := scanner.Text()
		list = append(list, Data{Key: strings.Split(s, " | ")[0], Value: s})
	}

	st := int(0)
	var end int
	end = len(list) - 1
	found := false

	for !found {

		if ((st+1 == end) || (st-1 == end) || (st == end)) && list[int(st)].Key != value {
			fmt.Printf(Red+"'%v' doesn't exist in the files\n"+Reset, value)
			break
		}
		if list[int((st+end)/2)].Key > value {
			end = (st + end) / 2
		} else if list[int((st+end)/2)].Key < value {
			st = (st + end) / 2
		} else if list[int((st+end)/2)].Key == value {
			found = true
			fmt.Println(Blue + list[int((st+end)/2)].Value + Reset)
		}

	}
}

type metadataEntry struct {
	file  int
	start string
	end   string
}
