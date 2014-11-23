package main

import(

	"io/ioutil"
	"encoding/json"
	"text/template"
	"github.com/fatih/color"
	"regexp"
	"os"
	"errors"
	"fmt"
)

const (
	dateFormat = "2006-01-02 15:04:05"
	configFile = "config.json"
)

var colorFunctions map[string]func(... interface{})string = map[string]func(... interface{})string{
	"red" : color.New(color.FgRed).SprintFunc(),
	"black" : color.New(color.FgBlack).SprintFunc(),
	"blue" : color.New(color.FgBlue).SprintFunc(),
	"green" : color.New(color.FgGreen).SprintFunc(),
	"yellow" : color.New(color.FgYellow).SprintFunc(),
	"magenta" : color.New(color.FgMagenta).SprintFunc(),
	"cyan" : color.New(color.FgCyan).SprintFunc(),
	"white" : color.New(color.FgWhite).SprintFunc(),
}

type ColorExpression struct {
	Color string `json:"color,omitempty"`
	Regexp string `json:"regexp,omitempty"`
	Colors []*ColorExpression `json:"colors,omitempty"`
	CompiledRegexp *regexp.Regexp `json:"-"`
}

func (c *ColorExpression) Colorize(text string) string {
	if c.CompiledRegexp == nil {
		c.CompiledRegexp = regexp.MustCompile(c.Regexp)
	}
	text = string(c.CompiledRegexp.ReplaceAllFunc([]byte(text), func(bytes []byte) []byte {
		if f,exists := colorFunctions[c.Color]; exists {
			text := f(string(bytes))
			if c.Colors == nil || len(c.Colors) == 0 {
				return []byte(text)
			}			
			for _, color := range c.Colors {
				text = color.Colorize(text)
			}
			return []byte(text)
		} else {
			return bytes;
		}
	}))



	return text
}

type Config struct {
	AccessKey 		string `json:"accessKey,omitempty"`
	SecretKey 		string `json:"secretKey,omitempty"`
	Endpoint  		string `json:"endpoint,omitempty"`
	Region    		string `json:"region,omitempty"`
	LogGroupName  *string  `json:"logGroupName,omitempty"`
	LogStreamName *string  `json:"logStreamName,omitempty"`
	Forward       *bool    `json:"stream,omitempty"`
	Match         *string  `json:"match,omitempty"`
	Flatten       *bool    `json:"flatten,omitempty"`
	Capture       *string  `json:"capture,omitempty"`
	ListGroups    *bool    `json:"listGroups,omitempty"`
	ListStreams   *bool		`json:"listStreams,omitempty"`
	FromDate	  *string   `json:"fromDate,omitempty"`
	ToDate		  *string   `json:"toDate,omitempty"`
	Colorize      *bool     `json:"colorize,omitempty"`
	Format        *string   `json:"format,omitempty"`
	Timestamp 	  *string   `json:"timestamp,omitempty"`
	FormatTemplate *template.Template `json:"-"`
	Config        *string   `json:"config,omitempty"`
	Colors        []*ColorExpression `json:"colors,omitempty"`
}

func (c *Config) ApplyColorize(text string) string {
	if c.Colors == nil || len(c.Colors) == 0 {
		return text
	}
	for _, color := range c.Colors {
		text = color.Colorize(text)
	}
	return text
}

func LoadConfig(filename string) (*Config, error) {

	if _, err := os.Stat(filename); os.IsNotExist(err) && filename == configFile {
		config := Config{AccessKey : "updateme", SecretKey : "updateme", Endpoint : "https://logs.us-east-1.amazonaws.com", Region : "us-east-1"}
		bytes, err := json.MarshalIndent(config, "", "    ")
		if err != nil {
			return nil, err
		}
		err = ioutil.WriteFile(configFile, bytes, 0644)
		if err != nil {
			return nil, err
		}
		return nil, errors.New("created config.json, please set your amazon aws credentials")
	}

	if bytes, err := ioutil.ReadFile(filename); err != nil {
		return nil, err
	} else {
		var config Config
		if err = json.Unmarshal(bytes, &config); err != nil {
			return nil, err
		} else {


			if len(config.AccessKey) == 0 || len(config.SecretKey) == 0 || config.AccessKey == "updateme" || config.SecretKey == "updateme" {
				return nil, errors.New(fmt.Sprintf("please set your amazon aws credentials in %s", filename))
			}


			return &config, nil
		}
	}
}

func Apply(target *Config, defaults *Config) *Config {
	targetJson, _ := json.Marshal(target)
	defaultJson, _ := json.Marshal(defaults)
	var targetMap map[string]interface{}
	json.Unmarshal(targetJson, &targetMap)
	var defaultMap map[string]interface{}
	json.Unmarshal(defaultJson, &defaultMap)
	for key, value := range defaultMap {
		targetMap[key] = value
	}
	targetJson, _ = json.Marshal(targetMap)
	var targetMapped Config
	json.Unmarshal(targetJson, &targetMapped)
	return &targetMapped
}