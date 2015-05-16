package config

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	Gotrap gotrapConfiguration `json:"gotrap"`
	Github GithubConfiguration `json:"github"`
	Amqp   AmqpConfiguration   `json:"amqp"`
	Gerrit GerritConfiguration `json:"gerrit"`
}

type gotrapConfiguration struct {
	Concurrent int `json:"concurrent"`
}

type GithubConfiguration struct {
	Username               string                    `json:"username"`
	APIToken               string                    `json:"api-token"`
	Organisation           string                    `json:"organisation"`
	Repository             string                    `json:"repository"`
	BranchPollingIntervall int                       `json:"branch-polling-intervall"`
	StatusPollingIntervall int                       `json:"status-polling-intervall"`
	PRTemplate             githubPullRequestTemplate `json:"pull-request"`
}

type githubPullRequestTemplate struct {
	Title string   `json:"title"`
	Body  []string `json:"body"`
}

type AmqpConfiguration struct {
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

type GerritConfiguration struct {
	URL            string                     `json:"url"`
	Username       string                     `json:"username"`
	Password       string                     `json:"password"`
	Projects       map[string]map[string]bool `json:"projects"`
	ExcludePattern []string                   `json:"exclude-pattern"`
	Comment        []string                   `json:"comment"`
}

func NewConfiguration(configFile *string) (*Configuration, error) {
	fileContent, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return nil, err
	}

	var config Configuration
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
