package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type gotrapConfiguration struct {
	Concurrent int `json:"concurrent"`
}

type githubPullRequestTemplate struct {
	Title string   `json:"title"`
	Body  []string `json:"body"`
}

type githubConfiguration struct {
	Username               string                    `json:"username"`
	APIToken               string                    `json:"api-token"`
	Organisation           string                    `json:"organisation"`
	Repository             string                    `json:"repository"`
	BranchPollingIntervall int                       `json:"branch-polling-intervall"`
	StatusPollingIntervall int                       `json:"status-polling-intervall"`
	PRTemplate             githubPullRequestTemplate `json:"pull-request"`
}

type amqpConfiguration struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	VHost      string `json:"vhost"`
	Exchange   string `json:"exchange"`
	Queue      string `json:"queue"`
	RoutingKey string `json:"routing-key"`
	Identifier string `json:"identifier"`
}

type gerritConfiguration struct {
	URL      string   `json:"url"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Comment  []string `json:"comment"`
}

type Configuration struct {
	Gotrap gotrapConfiguration `json:"gotrap"`
	Github githubConfiguration `json:"github"`
	Amqp   amqpConfiguration   `json:"amqp"`
	Gerrit gerritConfiguration `json:"gerrit"`
}

func NewConfiguration(configFile *string) *Configuration {
	fileContent, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal("Configuration file not found:", *configFile, err)
	}

	var config Configuration

	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		log.Fatal("JSON parsing failed:", *configFile, err)
	}

	return &config
}
