package main

import (
	"net/http"
	"fmt"
	"bufio"
	"os"
	"strings"
)

func translateToUrl(url string) string {
	url = strings.TrimSpace(url)
	newUrl := strings.Replace(url, " ", "+", -1)
	return "https://hooktube.com/results?search_query="+newUrl
}

func readInput() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter search term: ")
	text, _ := reader.ReadString('\n')
	return text
}
		

func main() {
	text := readInput()
	str := translateToUrl(text)
	fmt.Println(str)
}