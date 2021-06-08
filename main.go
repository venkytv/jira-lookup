package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/andygrunwald/go-jira"
	"github.com/keybase/go-keychain"
)

const (
	ApiTokenEnv = "JIRA_API_TOKEN"
	Site        = "https://mobiledgex.atlassian.net"
)

func getToken(username string) (string, error) {
	if token, ok := os.LookupEnv(ApiTokenEnv); ok {
		return token, nil
	}

	// Look up API token in keychain
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassInternetPassword)
	query.SetService(Site)
	query.SetAccount(username)
	query.SetReturnData(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		return "", err
	}
	nresults := len(results)
	switch {
	case nresults < 1:
		return "", fmt.Errorf("No keychain matches for account %s and service %s",
			username, Site)
	case nresults > 1:
		return "", fmt.Errorf("Too many results")
	default:
		return string(results[0].Data), nil
	}
}

func getClient(user string, token string) *jira.Client {
	tp := jira.BasicAuthTransport{
		Username: user,
		Password: token,
	}
	client, err := jira.NewClient(tp.Client(), "https://mobiledgex.atlassian.net")
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func userVal(user *jira.User) string {
	u := user.DisplayName
	if len(user.EmailAddress) > 0 {
		u = fmt.Sprintf("%s <%s>", u, user.EmailAddress)
	}
	return u
}

func printDetails(client *jira.Client, issueId string, field string) {
	issue, _, err := client.Issue.Get(issueId, nil)
	if err != nil {
		log.Fatal(err)
	}

	var resp string

	switch field {
	case "summary":
		resp = issue.Fields.Summary
	case "description":
		resp = issue.Fields.Description
	case "status":
		resp = issue.Fields.Status.Description
	case "assignee":
		resp = userVal(issue.Fields.Assignee)
	case "reporter":
		resp = userVal(issue.Fields.Reporter)
	default:
		resp = issueId + " " + issue.Fields.Summary
	}
	fmt.Println(resp)
}

func main() {
	var username, field string

	currUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&username, "username", currUser.Username, "Jira username")
	flag.StringVar(&field, "field", "", "Issue field to retrieve")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		log.Fatalf("usage: %s <issue-id>\n", os.Args[0])
	}
	issueId := args[0]

	apiToken, err := getToken(username)
	if err != nil {
		log.Fatal(err)
	}

	client := getClient(username, apiToken)
	printDetails(client, issueId, field)
}
