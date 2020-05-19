package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type StarredRepo struct {
	ID          int    `json:"id"`
	FullName    string `json:"full_name"`
	URL         string `json:"html_url"`
	Description string `json:"description"`
}

// GetRandGithubStars returns a random page of repositories that the provided
// username has starred on Github.
func GetRandGithubStars(username string) ([]StarredRepo, error) {
	res, err := http.Head(makeGithubURL(username))
	if err != nil {
		return nil, err
	}

	nPages, err := getNumberOfPages(res)
	if err != nil {
		return nil, err
	}

	var stars []StarredRepo
	var url string
	if nPages == 1 {
		// rand.Intn(1-1) will panic
		url = makePagedGithubURL(username, 1)
	} else {
		rand.Seed(time.Now().UnixNano())
		url = makePagedGithubURL(username, rand.Intn(nPages-1)+1)
	}

	res, err = http.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	err = json.Unmarshal(body, &stars)
	if err != nil {
		return nil, err
	}

	return stars, nil
}

func makeGithubURL(username string) string {
	return strings.Join([]string{"https://api.github.com/users", username, "starred"}, "/")
}

func makePagedGithubURL(username string, page int) string {
	return fmt.Sprintf("%s%s%d", makeGithubURL(username), "?page=", page)
}

func getNumberOfPages(res *http.Response) (int, error) {
	links := strings.Split(res.Header.Get("link"), ",")
	lastLink := links[len(links)-1]
	re := regexp.MustCompile(`\?page=(\d+)>`)
	matches := re.FindStringSubmatch(lastLink)
	pagestr := matches[len(matches)-1]
	pageint, err := strconv.Atoi(pagestr)
	if err != nil {
		return 0, err
	}
	return pageint, nil
}
