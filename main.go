package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
)

func readCSVFromUrl(url string) ([][]string, error) {
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

func main() {

	// client := github.NewClient(nil)

	// _, directoryContent, _, err := client.Repositories.GetContents(context.Background(), "CSSEGISandData", "COVID-19", "csse_covid_19_data/csse_covid_19_daily_reports", nil)

	// if err != nil {
	// 	panic(err)
	// }

	// for _, v := range directoryContent {
	// 	fmt.Println(v.GetType(), v.GetName(), v.GetDownloadURL(), v.GetSHA())
	// }
	url := "https://raw.githubusercontent.com/CSSEGISandData/COVID-19/master/csse_covid_19_data/csse_covid_19_daily_reports/03-16-2020.csv"

	resp, _ := readCSVFromUrl(url)

	for i := 0; i < len(resp); i++ {
		// fmt.Println(strings.Split(resp[i], ","))
		// fmt.Printf("%T", resp[i])
		fmt.Println(resp[i][1])
	}
	// fmt.Println(resp)

	// resp, err := http.Get("https://api.github.com/repos/CSSEGISandData/COVID-19/git/trees/e22872e7e9ea17b968386c79437a431ebec09d7d")

	// if err != nil {
	// 	panic(err)
	// }
	// defer resp.Body.Close()

	// // fmt.Println("Response status:", resp.Status)

	// b, err := ioutil.ReadAll(resp.Body)

	// if err != nil {
	// 	panic(err)
	// }

	// // fmt.Printf("%s", b)

	// var dat map[string]interface{}

	// if err := json.Unmarshal(b, &dat); err != nil {
	// 	panic(err)
	// }
	// fmt.Println(dat["tree"].([]interface{})[0].(string))

	// // for i := 0; i < len(); i++ {
	// // 	fmt.Println(i)
	// // }
}
