package main

import (
	code "1306170097/fileorg/project/Code"
	"bufio"
	"fmt"
	"os"
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

func main() {

	pwd, _ := os.Getwd() //project working directory

	selection := "n"
	var searchTerm string
	scanner := bufio.NewScanner(os.Stdin)

	//CLI menu
	for {
		if selection == "n" {
			fmt.Print("enter " + Yellow + "p" + Reset + " to process files, " +
				Yellow + "s" + Reset + " to search and " +
				Yellow + "e" + Reset + " to exit: ")
			scanner.Scan()
			selection = scanner.Text()
			fmt.Println()
		} else if selection == "p" {
			start := time.Now()
			code.Process(pwd)
			t := time.Since(start)
			sec := int(t / time.Second)
			mili := int(t/time.Millisecond) - sec*1000
			fmt.Printf(Magenta+"Total time: "+Cyan+"%v seconds %v milliseconds \n"+Reset, sec, mili)
			selection = "n"
			continue
		} else if selection == "s" {
			fmt.Print(Green + "enter the password to be searched: " + Reset)
			scanner.Scan()
			searchTerm = scanner.Text()
			searchTime := time.Now()
			code.Search(searchTerm, pwd)
			t := time.Since(searchTime)
			sec := int(t / time.Second)
			mili := int(t/time.Millisecond) - sec*1000
			fmt.Printf(Magenta+"Total search time:"+Cyan+" %v seconds %v milliseconds \n"+Reset, sec, mili)
			selection = "n"
			fmt.Println()
		} else if selection == "e" {
			break
		} else {
			selection = "n"
			continue
		}
	}
}
