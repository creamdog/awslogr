package main

import(	
	"fmt"
	"github.com/creamdog/goamz/logs"
	"time"
	"strings"
)

func listGroups(client *logs.CloudWatchLogs, config *Config) {
	groups, err := client.DescribeLogGroups(&logs.DescribeLogGroupsRequest{})
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	table := NewTable()
	table.AddRow("NAME", "CREATED", "STORED BYTES", "RETENTION IN DAYS")
	
	for _, group := range groups {
		table.AddRow(group.LogGroupName, 
			time.Unix(group.CreationTime/1000, 0).Format(*config.Timestamp), 
			fmt.Sprintf("%d", group.StoredBytes),
			fmt.Sprintf("%d", group.RetentionInDays))
	}

	fmt.Printf("\n")
	table.Print()
	fmt.Printf("\n")
}

func listStreams(client *logs.CloudWatchLogs, groupName string, config *Config) {
	streams, err := client.DescribeLogStreams(&logs.DescribeLogStreamsRequest{
		LogGroupName: groupName,
	})
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	table := NewTable()
	table.AddRow("NAME", "CREATED", "LAST EVENT TIMESTAMP", "STORED BYTES", "SEQUENCE TOKEN")
	
	for _, stream := range streams {
		table.AddRow(stream.LogStreamName, 
			time.Unix(stream.CreationTime/1000, 0).Format(*config.Timestamp), 
			time.Unix(stream.LastEventTimestamp/1000, 0).Format(*config.Timestamp), 
			fmt.Sprintf("%d", stream.StoredBytes),
			stream.UploadSequenceToken)
	}

	fmt.Printf("\n")
	table.Print()
	fmt.Printf("\n")
}

type EventStats struct {
	MinTimestamp int64
	MaxTimestamp int64
}

func listen(client *logs.CloudWatchLogs, config *Config, stats EventStats, groupName string, streamName string, initialDryRun bool, nextToken string, callback func([]*logs.Event), done func()) {


	request := logs.GetLogEventsRequest{
		LogGroupName:  groupName,
		LogStreamName: streamName,
		NextToken:     nextToken,
	}


	if !*config.Forward {

		start, err := time.ParseInLocation(dateFormat, *config.FromDate, time.Now().Location())
		if err != nil {
			fmt.Printf("%v\n", err)
			done()
			return
		}
		request.StartTime = start.UnixNano() / 1000000

		end, err := time.ParseInLocation(dateFormat, *config.ToDate, time.Now().Location())
		if err != nil {
			fmt.Printf("%v\n", err)
			done()
			return
		}
		request.EndTime = end.UnixNano() / 1000000

		request.StartFromHead = true
	}

	
	eventsReponse, err := client.GetLogEvents(&request)
	if err != nil {
		if strings.Index(err.Error(), "ThrottlingException") > 0 {
			fmt.Printf("%v\nSLEEPING 15 sec", err)
			time.Sleep(15 * time.Second)
		} else {
			fmt.Printf("%v\n", err)
			done()
			return
		}
	} else if *config.Forward {
		if !initialDryRun {
			callback(eventsReponse.Events)
		}			
		if eventsReponse.NextForwardToken == nextToken {
			time.Sleep(4 * time.Second)
		} else {
			time.Sleep(2 * time.Second)
		}
		nextToken = eventsReponse.NextForwardToken
	} else {

		if len(eventsReponse.Events) <= 0 {

			done()
			return
			//fmt.Printf("min: %s, max: %s\n", time.Unix(stats.MinTimestamp/1000, 0).Format(dateFormat), time.Unix(stats.MaxTimestamp/1000, 0).Format(dateFormat))
		}

		for _, event := range eventsReponse.Events {
			if stats.MaxTimestamp == 0 {
				stats = EventStats{event.Timestamp, event.Timestamp}
			}

			if(stats.MaxTimestamp < event.Timestamp) {
				stats.MaxTimestamp = event.Timestamp
			}

			if(stats.MinTimestamp > event.Timestamp) {
				stats.MinTimestamp = event.Timestamp
			}
		}

		callback(eventsReponse.Events)

		if eventsReponse.NextForwardToken == nextToken {
			done()
			return
		} else {
			time.Sleep(1 * time.Second)
		}
		nextToken = eventsReponse.NextForwardToken
	}

	

	listen(client, config, stats, groupName, streamName, false, nextToken, callback, done)
}
