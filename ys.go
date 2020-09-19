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

type issue struct {
	ID string
}

func getIssueID(host *string, token *string, task *string) (string, error) {
	v := url.Values{}
	v.Set("fields", "idReadable,id,summary")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/issues/?%s", *host, v.Encode()), nil)

	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", *token))
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

func setIssueState(host *string, token *string, issue *string, newState *string) error {
	v := url.Values{}
	v.Set("fields", "customFields(id,name,value(name))")

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/api/issues/%s?%s", *host, *issue, v.Encode()),
		bytes.NewBufferString(fmt.Sprintf(`{ "customFields":[ { "name": "State", "$type": "StateIssueCustomField", "value": { "name": "%s" } }]}`, *newState)))

	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", *token))
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

	if resp.StatusCode != 200 {
		return errors.New("Error")
	}

	return nil
}

func main() {
	taskID := flag.String("task", "", "task id")
	newState := flag.String("ns", "", "new state for task")
	host := flag.String("host", "", "host with proto like: https://youtracl.adnous.ru")
	token := flag.String("token", "", "YouTrack API token")

	flag.Parse()

	if *taskID == "" || *newState == "" || *host == "" || *token == "" {
		flag.Usage()
		os.Exit(200)
	}

	err := setIssueState(host, token, taskID, newState)

	if err != nil {
		log.Fatal(err)
		os.Exit(200)
	}

	log.Println("Success!")
}
