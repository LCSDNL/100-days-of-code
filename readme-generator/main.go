package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/IAPOLINARIO/100-days-of-code/readme-generator/structs"
	"github.com/olekukonko/tablewriter"
)

const (
	baseGithubURI    = "https://api.github.com"
	collaboratorsURI = "/repos/IAPOLINARIO/100-days-of-code/collaborators"
	commitsURI       = "/repos/IAPOLINARIO/100-days-of-code/commits"
	eventsURI        = "/repos/IAPOLINARIO/100-days-of-code/events"
	pullRequestsURI  = "/repos/IAPOLINARIO/100-days-of-code/pulls?state=closed"
)

func main() {
	token := os.Args[1]

	//getCollaboratorStats(token)
	//getCommits(token)
	PRs := getPullRequests(token)

	buildOutputResult(PRs)
}

func getGithubAPIResult(repoUrl string, token string) ([]byte, error) {
	// Create a token string by appending string access token
	var bearer = "token  " + token

	client := &http.Client{}
	req, _ := http.NewRequest("GET", repoUrl, nil)
	req.Header.Add("Authorization", bearer)

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func getCollaboratorStats(token string) {

	bodyBytes, _ := getGithubAPIResult(baseGithubURI+collaboratorsURI, token)
	// Convert response body to Contributors
	var contributors []structs.Contributor
	json.Unmarshal(bodyBytes, &contributors)

	fmt.Printf("Contributors on this repo:\n")
	for _, contr := range contributors {
		fmt.Printf("%v \n", contr.Login)
	}

	//fmt.Printf("API Response as struct %+v\n", todoStruct)
}

func getCommits(token string) {
	bodyBytes, _ := getGithubAPIResult(baseGithubURI+commitsURI, token)

	var commitBody []structs.Commit_Body
	json.Unmarshal(bodyBytes, &commitBody)

	fmt.Printf("Latest Commits:\n")
	for _, c := range commitBody {
		fmt.Printf("Author: %v\n", c.Author.Login)

		getcommit(token, c.Sha)
	}
	//bodyString := string(bodyBytes)
	//fmt.Println("API Response as String:\n" + bodyString)

}

func getcommit(token string, id string) {
	commitURI := baseGithubURI + commitsURI + "/" + id
	bodyBytes, _ := getGithubAPIResult(commitURI, token)

	var commit structs.Commit_Body
	json.Unmarshal(bodyBytes, &commit)

	//fmt.Printf("Files changed in this commit: %v, Total changes: %v\n", len(commit.Files), commit.Stats.Total)
	for _, f := range commit.Files {
		//   \b(\w*day-\w*)\b     			-> match whole word

		rg := regexp.MustCompile("day-\\w*")
		match := rg.FindStringSubmatch(f.Filename)

		if len(match) > 0 {
			fmt.Printf("Challenge of day: %v was solved \n", match[0])
		}
	}

}

func getEvents(token string) {
	bodyBytes, _ := getGithubAPIResult(baseGithubURI+eventsURI, token)
	fmt.Printf("Latest Events:\n")

	bodyString := string(bodyBytes)
	fmt.Println("API Response as String:\n" + bodyString)

}

func buildOutputResult(PRs []structs.PullRequest) {
	req := make(map[string][]string)
	for i := 0; i < len(PRs); i++ {
		challengesDone := ""
		totalPoints := 0
		userHasDoneLabel := false

		for _, l := range PRs[i].Labels {
			if l.Name == "done" {
				rg := regexp.MustCompile("day-\\w*")
				match := rg.FindStringSubmatch(PRs[i].Title)

				if len(match) > 0 {
					userHasDoneLabel = true
					if !strings.Contains(challengesDone, match[0]) {
						challengesDone = match[0]
						totalPoints += 100
					}
				}
			}
		}

		if userHasDoneLabel {
			if len(req[PRs[i].User.Login]) > 0 {
				currentUserValue := req[PRs[i].User.Login]
				challengesCombined := []string{currentUserValue[0], challengesDone}
				currentTotalScore, _ := strconv.Atoi(currentUserValue[1])
				newScore := totalPoints + currentTotalScore
				req[PRs[i].User.Login] = []string{strings.Join(challengesCombined, ","), strconv.Itoa(newScore)}

			} else {
				req[PRs[i].User.Login] = []string{challengesDone, strconv.Itoa(totalPoints)}
			}
		}

	}

	sortedMap := SortMapByValue(req)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Ranking", "Contributor", "Challenges", "Total Points"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	for k, v := range sortedMap {
		result := []string{v[0], k, v[1], v[2]}
		table.Append(result)
	}

	table.Render()
}

func SortMapByValue(Map map[string][]string) map[string][]string {

	resultMap := make(map[string][]string)
	// used to switch key and value
	hack := map[float64][]string{}
	hackkeys := []float64{}

	for key, val := range Map {
		NewKey, _ := strconv.ParseFloat(val[1], 64)
		NewValue := []string{key, val[0]}

		if len(hack[NewKey]) > 0 {
			NewKey += 0.00001
		}

		hackkeys = append(hackkeys, NewKey)

		hack[NewKey] = NewValue
	}

	sort.Slice(hackkeys, func(i, j int) bool {
		return hackkeys[i] >= hackkeys[j]
	})

	for i := 0; i < len(hackkeys); i++ {
		key := hack[hackkeys[i]][0]

		score := fmt.Sprintf("%.0f", hackkeys[i])
		value := []string{strconv.Itoa(i + 1), hack[hackkeys[i]][1], score}
		resultMap[key] = value
	}

	return resultMap
}

func getPullRequests(token string) (PR []structs.PullRequest) {
	bodyBytes, _ := getGithubAPIResult(baseGithubURI+pullRequestsURI, token)

	fmt.Println(baseGithubURI + pullRequestsURI)
	var PRs []structs.PullRequest
	json.Unmarshal(bodyBytes, &PRs)

	return PRs
	//fmt.Printf("These are the pull requests approved and done:\n")
}
