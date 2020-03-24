# covid-chat-notify

Simple notification for [Slack](https://slack.com)/[Rocketchat](https://rocket.chat) about COVID-19 stats for last day.

Based on [2019 Novel Coronavirus COVID-19 (2019-nCoV) Data Repository by Johns Hopkins CSSE](https://github.com/CSSEGISandData/COVID-19)

Install:
- Build for your OS/ARCH (`Example for Ubuntu: GOOS=linux GOARCH=amd64 go build -o covidnotifcation main.go`)
- Configure `config.yaml`
- Run it `./covidnotifcation` (or add it to crontab)