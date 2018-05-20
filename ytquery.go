package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DexterLB/mpvipc"
	"github.com/fatih/color"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type PlayerMessage int

const (
	PausePlayback PlayerMessage = 0
	StopPlayback  PlayerMessage = 1
	SkipPlayback  PlayerMessage = 2
)

var (
	YtqueryConnectionError       = errors.New("Could not connect to hooktube api.")
	YtqueryUnsuportedOptionError = errors.New(red("Unsupported option."))
)

// Color funtion definitions
var (
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
)



// Basic utilty functions
////////////////////////////////////////////////

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
	query := fmt.Sprintf(" Play (%s) %s ", blue("0-49"), yellow("::"))


	sel, err := strconv.Atoi(readInput(query))
	if err != nil {
		return 0, err
	}
	return sel, nil
}


type Queue struct {
	Fifo chan Item
	List []Item
}

func (q *Queue) WriteToFile() {
}

func (q *Queue) listQueue() {
	fmt.Println()
	fmt.Printf("    %s Queue %s\n", yellow("::"), yellow("::"))
	for i, item := range q.List {
		fmt.Printf("   %2s %s %s\n", blue(i+1), yellow("::"), item.giveTitle())
	}
	fmt.Println()
}

func (q *Queue) popSong() Item {
	// Remove first element from slice
	if len(q.List)-1 > 0 {
		tmp := make([]Item, len(q.List)-1)
		copy(tmp, q.List[1:])
		q.List = tmp
	} else {
		q.List = make([]Item, 0)
	}

	// TODO Add select here not to hang the program on an empty
	// list skip
	return <-q.Fifo
}

//Mpv struct related
////////////////////////////////////////////////

type MPV struct {
	Conn   mpvipc.Connection
	Result Results
	Queue  Queue
	Paused bool
}

func newMpvPlayer(socket string) MPV {
	return MPV{Conn: *mpvipc.NewConnection(socket), Queue: Queue{Fifo: make(chan Item, 30), List: make([]Item, 0)}}
}

func (m *MPV) nextSong() {
	item := m.Queue.popSong()
	m.Conn.Call("loadfile", item.giveLink(), "append-play")
}

func (mpv *MPV) makeQuery() error {
	// Read search
	query := fmt.Sprintf("\n Enter search term %s ", yellow("::"))
	text := readInput(query)
	str := translateToUrl(text)

	//Server query
	resp, err := http.Get(str)
	if err != nil {
		return YtqueryConnectionError
	}
	defer resp.Body.Close()

	// Convert body to []byte
	body, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(body, &mpv.Result)

	return nil
}

func (m *MPV) Play(messages <-chan PlayerMessage) {

	m.Conn.Open()
	events := make(chan *mpvipc.Event)
	stop := make(chan struct{})
	go m.Conn.ListenForEvents(events, stop)

	for {
		select {
		case message := <-messages:
			switch message {
			case PausePlayback:
				m.Conn.Call("cycle", "pause")
				m.Paused = !m.Paused

			case StopPlayback:
				m.Conn.Call("playlist-remove", "current")

			case SkipPlayback:
				m.nextSong()
				m.Conn.Call("playlist-next")
			}

		case event := <-events:
			if event.Name == "end-file" {
				m.nextSong()
			}
		}
	}
}

func (m *MPV) addToQueue(id int) {
	m.Queue.List = append(m.Queue.List, m.Result.Items[id])
	m.Queue.Fifo <- m.Result.Items[id]
}

// Results
////////////////////////////////////////////////

type Results struct {
	Items []Item `json:"items"`
}

// Loop over items and print titles
func (r *Results) printResults() {
	for i, item := range r.Items {
		fmt.Printf("   %2s %s %s\n", blue(i), yellow("::"), item.giveTitle())
	}
	fmt.Println()
}

// Item
////////////////////////////////////////////////

type Item struct {
	Id      map[string]string `json:"id"`
	Snippet map[string]string `json:"snippet"`
}

// Wrappers around the single item in a json response
func (i *Item) giveId() string {
	return i.Id["videoId"]
}

func (i *Item) giveLink() string {
	return "https://hooktube.com/watch?v=" + i.giveId()
}

func (i *Item) giveTitle() string {
	return i.Snippet["title"]
}

////////////////////////////////////////////////

func main() {

	messages := make(chan PlayerMessage)

	// Initiate pipe
	var player MPV
	switch os := runtime.GOOS; os {
	case "windows":
		player = newMpvPlayer("\\\\.\\pipe\\ytquery_socket")
	default:
		player = newMpvPlayer("/tmp/ytquery_socket")
	}
	defer player.Conn.Close()

	go player.Play(messages)

	query := fmt.Sprintf(" [%s]ist | [%s]dd | [%s]arch | [%s]ip | [%s]lay/ause | [%s]top | [%s]uit %s ", blue("l"),
		blue("a"), cyan("se"), green("sk"), yellow("p"), yellow("s"), red("q"), yellow("::"))

mainLoop:
	for {

		switch i := readInput(query); i {
		case "l":
			player.Queue.listQueue()
		case "a":
			sel, err := chooseVideo()

			if err != nil {
				fmt.Println(err)
			}

			player.addToQueue(sel)

		case "se":
			err := player.makeQuery()
			if err != nil {
				fmt.Println("Error: ", err)
			}
			player.Result.printResults()
		case "p":
			messages <- PausePlayback
		case "s":
			messages <- StopPlayback
		case "sk":
			messages <- SkipPlayback
		case "q":
			break mainLoop
		}
	}
}
