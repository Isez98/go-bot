package main

import(
	"os"
	"fmt"
	"log"
	"github.com/slack-go/slack"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
	"bytes"
	"strings"
)

var (
	slackClient *slack.Client
)

type SearchResults struct {
	Items []Item `json:items` 
}

type Item struct {
	Link string `json:link`
	Snippet string `json:snippet`
	Title string `json:title`
}

type TextInfo struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Block struct {
	Type string `json:"type"`
	Text TextInfo `json:"text"`
	BlockId string `json:"block_id"`
}

type Payload struct {
	Blocks []Block `json:"blocks"`
}

func main() {
	slackClient = slack.New(os.Getenv("SLACK_ACCESS_TOKEN"))
	rtm := slackClient.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if len(ev.BotID) == 0 {
				go handleMessage(ev)
			}
		}
	}
}

func handleMessage(ev *slack.MessageEvent) {
	fmt.Printf("%v\n", ev)
	structure := searchAnswer(ev)
	
	jsonMessage := dataBinding(structure)

	replyToUser(jsonMessage)
	options := slack.MsgOptionText("Hello there",false)
	slackClient.PostMessage(ev.Channel,options)
}

func replyToUser(jsonMessage []byte) {	

	resp, err := http.Post(os.Getenv("WEB_HOOK"), "application/json", bytes.NewBuffer(jsonMessage))
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, getErr := ioutil.ReadAll(resp.Body)
	if getErr != nil {
		log.Fatalln(getErr)
	}
	log.Println(body)
}

func searchAnswer(ev *slack.MessageEvent) SearchResults {
		url := "https://www.googleapis.com/customsearch/v1?key=AIzaSyD8QNzBdjzt3ZNEbGTz4P1rSAnvDPtbrUU&cx=005033773481765961543:gti8czyzyrw&num=3&q=golang"
		googleClient := http.Client{
			Timeout: time.Second * 3, // Maximum of 3 secs
		}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("User-Agent", "Isacc Hernandez")
		res, getErr := googleClient.Do(req)
		if getErr != nil {
			log.Fatal(getErr)
		}
		if res.Body != nil {
				defer res.Body.Close()
			}
		body, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			log.Fatal(readErr)
		}
		value := apiMessage(body)
		return value
}

func apiMessage(jsonRaw []byte) SearchResults {
	structure := SearchResults{}
	jsonErr := json.Unmarshal(jsonRaw, &structure)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
  return structure
}

func dataBinding(data SearchResults) []byte {
	payload := new(Payload)

	for i:=0; i<3; i++ {
		item := data.Items[i]
		textBlock := fmt.Sprintf("*<%s|%s>*\n>_%s_", item.Link, item.Title, strings.Replace(item.Snippet, "\n", " ", -1))
		block := Block{
			Type: "section",   
			Text: TextInfo{"mrkdwn", textBlock},
			BlockId: fmt.Sprintf("text%v", i),
		}
		payload.Blocks = append(payload.Blocks, block)
	}

	jsonMessage, recieveErr := json.Marshal(payload)
	if recieveErr != nil {
		log.Fatalln(recieveErr)
	}
	return jsonMessage
}