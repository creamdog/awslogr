awslogr
=======

Command line application for accessing Amazon AWS CloudWatch logs

currently supporting:

- streaming and visualizing event data
- querying / searching event logs data
- advanced formatting / output
- syntax highlightning

## How to build and install

* `$ go get github.com/creamdog/awslogr`
* `$ go build github.com/creamdog/awslogr`
* `$ go install github.com/creamdog/awslogr`

## Usage

* for command options run `awslogr -h`

the first time awslogr is run it will create a configuration file names `config.json`, make sure to fill in your Amazon AWS credentials, [region and CloudWatch endpoint](http://docs.aws.amazon.com/general/latest/gr/rande.html#cw_region)

```
{
	"accessKey" : "g0fd87gd98g7fdg987",
	"secretKey" : "KJjhfdu6d9asd7a8d76adas76d8sa8",
	"region" : "us-east-1",
	"endpoint" : "https://logs.us-east-1.amazonaws.com",
	"logGroupName" : "nodejs-cluster"
}
```

###example query
`> awslogr -groupName "node-cluster" -match "(?i)exception" -fromDate "2014-11-21 00:00:00" -toDate "2014-11-22 00:00:00" -capture "^[^\n]{1,200}" -format "<<{{.Timestamp}}>> {{.Message}}{{.Newline}}"`

*switches*

- groupName: "the name of the Amazon AWS CloudWatch Log Group, example matches the word 'exception', case insensitive"
- match: "regular expression to match against (perl style)"
- fromDate: "time to start searching from (local time)"
- toDate: "time up until to search to (local time)"
- capture: "regular expression used to capture or trim output, example matches first 200 characters or until new-line"
- format: "controls the format logs are outputted using {{.Timestamp}}, {{.Message}} and {{.Newline}}"

*output*

```
<<2014-11-21 00:01:23>> SyntaxException, something went wrong with something in the class blah @ blah
<<2014-11-21 00:15:15>> wow, this is some string with the word "exception" in it
<<2014-11-21 00:16:33>> EXCEPTION WOW FAIL
```

###example, stream events into sql file

`awslogr -groupName "node-cluster" -match . -stream -format "INSERT INTO events(timestamp, message) VALUES('{{.Timestamp}}','{{.Timestamp}}');{{.Newline}}" > capture.sql`

- groupName: "the name of the Amazon AWS CloudWatch Log Group, example matches the word 'exception', case insensitive"
- match: "regular expression to match against (perl style), example matches everything"
- stream: switch that tells awslogr to capture new events from the HEAD of all logstreams for the current log group.
- format: "controls the format logs are outputted using {{.Timestamp}}, {{.Message}} and {{.Newline}}"

*output into capture.sql*
```
INSERT INTO events(timestamp, message) VALUES('2014-11-21 00:01:23', 'SyntaxException, something went wrong with something in the class blah @ blah');
INSERT INTO events(timestamp, message) VALUES('2014-11-21 00:01:33', 'wow, this is some string with the word "exception" in it');
INSERT INTO events(timestamp, message) VALUES('2014-11-21 00:05:10', 'EXCEPTION WOW FAIL');
```
