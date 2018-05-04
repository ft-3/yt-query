package main

import (
	"net/http"
	"fmt"
	"bufio"
	"os"
	"strings"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"os/exec"
)


// Data type to store results from api query
type Results struct {
	Items []Item `json:"items"`
}

type Item struct {
	Id      map[string]string `json:"id"`
	Snippet map[string]string `json:"snippet"`
}

// Wrappers to extract id ex.: -tqZZmF5wlI or video title
func (r *Results) giveId(sel int) string {
	return r.Items[sel].Id["videoId"]
}

func (r *Results) giveTitle(sel int) string {
	return r.Items[sel].Snippet["title"]
}

// Loop over items and print titles
func (r *Results) printResults() {
	for i, item := range r.Items {
		fmt.Printf("%2d <--> %s\n", i, item.Snippet["title"])
	}
}

func makeQuery() (Results, error) {
	// Read search
	text := readInput("Enter search term: ")
	str := translateToUrl(text)

	//Server query
	resp, err := http.Get(str)
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer resp.Body.Close()

	// Convert body to []byte
	body, _ := ioutil.ReadAll(resp.Body)

	var result Results
	json.Unmarshal(body, &result)

	return result, nil
}

// Removes spaces from entered query, substitutes them with %20, and returns
// link to hooktube api with read query.
func translateToUrl(url string) string {
	url = strings.TrimSpace(url)
	newUrl := strings.Replace(url, " ", "%20", -1)
	return "https://hooktube.com/api?mode=search&q=" + newUrl
}


// Wrapper function to read user query with specified label
func readInput(query string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(query)
	text, _ := reader.ReadString('\n')
	return strings.Trim(text, " \n")
}


// Returns the user chosen video as int
func chooseVideo() (int, error) {
	selection := readInput("Id of video to play (0 - 49): ")
	if sel, err := strconv.Atoi(selection); err == nil {
		return sel, nil
	} else {
	return 0, err
	}
}


func idToLink(id string) string {
	return "https://hooktube.com/watch?v=" + id
}



func main() {
	result, _ := makeQuery()
	result.printResults()

	for {
		sel, err := chooseVideo()
		if err != nil {
			return 
		}
		id := result.giveId(sel)
		link := idToLink(id)
		title := result.giveTitle(sel)

		fmt.Println("Playing:", title, "<", link, ">")
		exec.Command("mpv", "--no-video", link).Run()
	}
}