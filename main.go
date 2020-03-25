package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"gopkg.in/yaml.v3"
)

// Row comment
type Row struct {
	province   string
	country    string
	confirmed  int
	deaths     int
	recovered  int
	lastUpdate string
}

// Messenger comment
type Messenger struct {
	Active      bool
	URL         string
	UserID      string `yaml:"user-id"`
	ChannelName string `yaml:"channel-name"`
	Token       string
}

// Config comment
type Config struct {
	Locations  []string
	Rocketchat Messenger
	Slack      Messenger
}

var lastFileURL, lastFileName, key, findKey string

func readCSVFromURL(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	// reader.Comma = ','
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func addData() {

}

func main() {

	t, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	config := Config{}

	err = yaml.Unmarshal(t, &config)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	client := github.NewClient(nil)

	_, directoryContent, _, err := client.Repositories.GetContents(context.Background(), "CSSEGISandData", "COVID-19", "csse_covid_19_data/csse_covid_19_daily_reports", nil)

	if err != nil {
		panic(err)
	}

	for _, v := range directoryContent {
		if strings.HasSuffix(v.GetName(), "csv") {
			lastFileURL = v.GetDownloadURL()
			lastFileName = v.GetName()
		}
	}

	lastFileDate, _ := time.Parse(
		"01-02-2006.csv",
		lastFileName)

	resp, _ := readCSVFromURL(lastFileURL)

	stats := make([]Row, 0)

	fullStats := map[string]Row{}

	// find need keys
	keys := map[string]int{
		"province": 0,
		"country":  0,
		"conf":     0,
		"death":    0,
		"recov":    0,
		"last_upd": 0,
	}
	for i, v := range resp[0] {
		researchString := strings.ToLower(v)
		if strings.Contains(researchString, "province") {
			keys["province"] = i
		}
		if strings.Contains(researchString, "country") {
			keys["country"] = i
		}
		if strings.Contains(researchString, "confirmed") {
			keys["conf"] = i
		}
		if strings.Contains(researchString, "deaths") {
			keys["death"] = i
		}
		if strings.Contains(researchString, "recovered") {
			keys["recov"] = i
		}
		if strings.Contains(researchString, "update") {
			keys["last_upd"] = i
		}
	}

	for i := 0; i < len(resp); i++ {
		key = ""

		confirmed, _ := strconv.Atoi(resp[i][keys["conf"]])
		deaths, _ := strconv.Atoi(resp[i][keys["death"]])
		recovered, _ := strconv.Atoi(resp[i][keys["recov"]])
		country := strings.ToLower(resp[i][keys["country"]])

		t1, _ := time.Parse(
			time.RFC3339,
			resp[i][4]+"Z")

		stats = append(stats, Row{
			province:   resp[i][keys["province"]],
			country:    resp[i][keys["country"]],
			confirmed:  confirmed,
			deaths:     deaths,
			recovered:  recovered,
			lastUpdate: t1.Format("2006-01-02 15:04:05"),
		})

		if resp[i][2] == "" {
			key = country
		} else {
			key = strings.ToLower(resp[i][keys["province"]])

			if _, err := fullStats[country]; err == false {
				fullStats[country] = Row{}
			}

			if thisRow, ok := fullStats[country]; ok {
				thisRow.country = resp[i][keys["country"]]
				thisRow.confirmed = thisRow.confirmed + confirmed
				thisRow.deaths = thisRow.deaths + deaths
				thisRow.recovered = thisRow.recovered + recovered
				thisRow.lastUpdate = t1.Format("2006-01-02 15:04:05")
				fullStats[country] = thisRow
			}
		}

		if _, err := fullStats[key]; err == false {
			fullStats[key] = Row{}
		}

		if thisRow, ok := fullStats[key]; ok {
			thisRow.province = resp[i][keys["province"]]
			thisRow.country = resp[i][keys["country"]]
			thisRow.confirmed = thisRow.confirmed + confirmed
			thisRow.deaths = thisRow.deaths + deaths
			thisRow.recovered = thisRow.recovered + recovered
			thisRow.lastUpdate = t1.Format("2006-01-02 15:04:05")
			fullStats[key] = thisRow
		}
	}

	messageText := "Date: " + lastFileDate.Format("02.01.2006") + "\n"

	if len(config.Locations) < 1 {
		panic("Invalid config. Set `locations` field")
	}

	for _, v := range config.Locations {
		findKey = strings.ToLower(v)
		if _, err := fullStats[findKey]; err == false {
			messageText = messageText + "Location `" + v + "` not found\n"
			continue
		}

		messageText = messageText + "Location `" + v + "` statistics: " +
			"Confirmed: " + strconv.Itoa(fullStats[findKey].confirmed) + "; " +
			"Deaths: " + strconv.Itoa(fullStats[findKey].deaths) + "; " +
			"Recovered: " + strconv.Itoa(fullStats[findKey].recovered) + "\n"
	}

	message := map[string]interface{}{
		"text":    messageText,
		"channel": "",
	}

	clientReq := &http.Client{}

	// ROCKET
	if config.Rocketchat.Active == true {
		message["channel"] = "#" + config.Rocketchat.ChannelName

		headersRocket := map[string]string{
			"X-Auth-Token": config.Rocketchat.Token,
			"X-User-Id":    config.Rocketchat.UserID,
		}

		send(clientReq, config.Rocketchat.URL, message, headersRocket)
	}

	// SLACK
	if config.Slack.Active == true {
		bearer := "Bearer " + config.Slack.Token
		slackChannel := "#" + config.Slack.ChannelName
		message["channel"] = slackChannel

		headersSlack := map[string]string{
			"Authorization": bearer,
			"Content-Type":  "application/json; charset=utf8",
		}

		send(clientReq, config.Slack.URL, message, headersSlack)
	}

}

func send(client *http.Client, URL string, sendData map[string]interface{}, headers map[string]string) {

	bytesRepresentation, err := json.Marshal(sendData)
	if err != nil {
		log.Fatalln(err)
	}

	req, _ := http.NewRequest("POST", URL, bytes.NewBuffer(bytesRepresentation))
	for i, v := range headers {
		req.Header.Add(i, v)
	}

	res, _ := client.Do(req)
	body, err := ioutil.ReadAll(res.Body)
	log.Println(string(body))

	res.Body.Close()
}
