package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/creamdog/goamz/logs"
	"github.com/crowdmob/goamz/aws"
	"github.com/fatih/color"
	"regexp"
	"strings"
	"text/template"
	"time"
)

func main() {

	defer color.Unset()

	red := color.New(color.FgRed).SprintFunc()

	config, err := LoadConfig(configFile)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	flags := &Config{
		LogGroupName:  flag.String("groupName", "", "name of log group"),
		LogStreamName: flag.String("streamName", "", "name of log stream"),
		Forward:       flag.Bool("stream", false, "stream new events, indefinetly"),
		Match:         flag.String("match", "(?is).+", "match against <regexp>"),
		Flatten:       flag.Bool("flatten", false, "replace newlines with spaces before capture"),
		Capture:       flag.String("capture", "(?is).+", "capture using <regexp>"),
		ListGroups:    flag.Bool("listGroups", false, "list log groups"),
		ListStreams:   flag.Bool("listStreams", false, "list log group streams"),
		FromDate:      flag.String("fromDate", time.Now().Add(-2*time.Hour).Format(dateFormat), "from date"),
		ToDate:        flag.String("toDate", time.Now().Format(dateFormat), "to date"),
		Colorize:      flag.Bool("colorize", false, "colorize"),
		Format:        flag.String("format", "[{{.Timestamp}}] {{.Message}}{{.Newline}}", "output format"),
		Timestamp:     flag.String("timestamp", dateFormat, "timestamp format"),
		Config:        flag.String("config", configFile, "config file to load"),
	}
	flag.Parse()

	flags = Apply(flags, config)

	flags = Apply(config, flags)

	if *flags.Config != configFile {
		config, err = LoadConfig(*flags.Config)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		flags = Apply(flags, config)
	}

	flags.FormatTemplate = template.Must(template.New("format").Parse(*flags.Format))

	auth := aws.Auth{AccessKey: config.AccessKey, SecretKey: config.SecretKey}
	client, err := logs.New(auth, config.Endpoint, config.Region)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	if *flags.ListGroups {
		listGroups(client, flags)
		return
	}

	if *flags.ListStreams {
		listStreams(client, *flags.LogGroupName, flags)
		return
	}

	matchRegexp := regexp.MustCompile(*flags.Match)
	captureRegexp := regexp.MustCompile(*flags.Capture)

	streams, err := client.DescribeLogStreams(&logs.DescribeLogStreamsRequest{
		LogGroupName: *flags.LogGroupName,
	})
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Printf("streams: %d ==> %q\n", len(streams), streams)

	workers := make(chan int, len(streams))
	for _, stream := range streams {

		go listen(client, flags, EventStats{0, 0}, *flags.LogGroupName, stream.LogStreamName, true, "", func(events []*logs.Event) {

			defer color.Unset()

			events = filter(events, func(event *logs.Event) bool {
				return matchRegexp.MatchString(event.Message)
			})

			if *flags.Flatten {
				events = transform(events, func(event *logs.Event) *logs.Event {
					event.Message = strings.Replace(event.Message, "\n", " ", -1)
					return event
				})
			}

			events = transform(events, func(event *logs.Event) *logs.Event {
				event.Message = captureRegexp.FindString(event.Message)
				return event
			})

			for _, event := range events {

				time := time.Unix(event.Timestamp/1000, 0).Format(*flags.Timestamp)
				text := event.Message
				if *flags.Colorize {
					time = red(time)
					text = config.ApplyColorize(text)
				}
				var doc bytes.Buffer
				flags.FormatTemplate.Execute(&doc, struct {
					Timestamp string
					Message   string
					Newline   string
				}{
					time,
					text,
					"\n",
				})
				fmt.Printf("%s", doc.String())
				color.Unset()
			}

		}, func() {
			workers <- 1
		})
	}
	for len(workers) < len(streams) {
		time.Sleep(2 * time.Second)
	}
}
