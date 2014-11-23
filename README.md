awslogr
=======

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
