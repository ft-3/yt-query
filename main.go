package main

import (
	"net/http"
	"fmt"
	"bufio"
	"os"
	"strings"
	"io/ioutil"
	"encoding/json"
)


// Data type to store results from api query
type Results struct {
	Items     []Item           `json:"items"`
	Id        map[string]string `json:"id"`
	Snippet   map[string]string `json:"snippet"`
}

type Item struct {
	Id      map[string]string      `json:"id"`
	Snippet map[string]string `json:"snippet"`
}

//func (i *Items) UnmarshalJSON(b []byte) (err error) {
	

func translateToUrl(url string) string {
	url = strings.TrimSpace(url)
	newUrl := strings.Replace(url, " ", "%20", -1)
	return "https://hooktube.com/api?mode=search&q=" + newUrl
}

func readInput() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter search term: ")
	text, _ := reader.ReadString('\n')
	return text
}

// Have to figure out a way to pull things out of the items field in the response.
// Thinking about doing a json.RawMessage type, and making my own decoder for it


func main() {
	text := readInput()
	str := translateToUrl(text)
	resp, err := http.Get(str)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Close the body when function ends
	defer resp.Body.Close()

	// Convert body to []byte
	body, _ := ioutil.ReadAll(resp.Body)

	// DEBUG
	reader := bufio.NewReader(os.Stdin)

	var structured Results
	//var result interface{}
	//json.Unmarshal(body, &result)

	json.Unmarshal(body, &structured)

	for i, item := range structured.Items {
		fmt.Println(">>", i, "<<")
		fmt.Println("ID -->", item.Id)
		fmt.Println("SNIPPET -->", item.Snippet)
		reader.ReadString('\n')
	}

	//m := structured.Items.(map[string]interface{})
/*
	fmt.Println(structured)
	fmt.Println(structured.Id)
	fmt.Println(structured.Snippet)
	for k, v := range m {
		switch vv := v.(type) {
		case string:
			fmt.Println(k, "is string")

		case float64:
			fmt.Println(k, "is float")

		case []interface{}:
			fmt.Println(k, "is an array")
			for i, u := range vv {
				fmt.Println(i, u)
				reader.ReadString('\n')
			}
		default:
			fmt.Println(k, "dunno m8")
		}
		reader.ReadString('\n')
	}
*/
}