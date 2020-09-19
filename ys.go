package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"gopkg.in/yaml.v2"
)

type config struct {
	Youtrack struct {
		Host    string
		Token   string
		Project string
		Prefix  string
	}
}

type issue struct {
	ID string
}

func readConfig(fileName string) (*config, error) {

	yamlFile, err := ioutil.ReadFile(fileName)

	if err != nil {
		return nil, err
	}

	conf := config{}

	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func getIssueID(config *config, task string) (string, error) {
	v := url.Values{}
	v.Set("fields", "idReadable,id,summary")
	v.Set("query", fmt.Sprintf("project:%s #%s", config.Youtrack.Project, task))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/issues/?%s", config.Youtrack.Host, v.Encode()), nil)

	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Youtrack.Token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accepted", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New("Issue not found")
	}

	bodyText, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var issues []issue

	err = yaml.Unmarshal(bodyText, &issues)
	if err != nil {
		return "", err
	}

	if len(issues) != 1 {
		return "", errors.New("Issue not found")
	}

	return issues[0].ID, nil
}

func setIssueState(config *config, issue *string, newState *string) error {
	v := url.Values{}
	v.Set("fields", "customFields(id,name,value(name))")

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/api/issues/%s?%s", config.Youtrack.Host, *issue, v.Encode()),
		bytes.NewBufferString(fmt.Sprintf(`{ "customFields":[ { "name": "State", "$type": "StateIssueCustomField", "value": { "name": "%s" } }]}`, *newState)))

	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Youtrack.Token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accepted", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	// bodyText, err := ioutil.ReadAll(resp.Body)

	// if err != nil {
	// 	return err
	// }

	if resp.StatusCode != 200 {
		return errors.New("Issue not found")
	}

	// log.Println(string(bodyText))

	return nil
}

func main() {
	taskPtr := flag.String("task", "", "task id")
	toStatePtr := flag.String("ns", "", "new state for task")
	configPtr := flag.String("config", "config.yml", "config file")

	flag.Parse()

	if *taskPtr == "" || *toStatePtr == "" {
		flag.Usage()
		os.Exit(200)
	}

	conf, err := readConfig(*configPtr)

	if err != nil {
		log.Fatal(err)
		os.Exit(200)
	}

	_, err = getIssueID(conf, *taskPtr)

	if err != nil {
		log.Fatal(err)
		os.Exit(200)
	}

	err = setIssueState(conf, taskPtr, toStatePtr)

	if err != nil {
		log.Fatal(err)
		os.Exit(200)
	}

	log.Println("Success!")
}
