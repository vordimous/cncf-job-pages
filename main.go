package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Item struct {
	Name        string `yaml:"name"`
	HomepageURL string `yaml:"homepage_url"`
	RepoURL     string `yaml:"repo_url,omitempty"`
}
type Subcategory struct {
	Name  string `yaml:"name"`
	Items []Item `yaml:"items"`
}
type Category struct {
	Name          string        `yaml:"name"`
	Subcategories []Subcategory `yaml:"subcategories"`
}

type cncf struct {
	Landscape []Category `yaml:"landscape"`
}

func main() {
	url := "https://raw.githubusercontent.com/cncf/landscape/refs/heads/master/landscape.yml"

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var data cncf
	err = yaml.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	urls := []string{}
	for _, l := range data.Landscape {
		for _, s := range l.Subcategories {
			for _, i := range s.Items {
				if !strings.HasSuffix(i.HomepageURL, "/") {
					i.HomepageURL = i.HomepageURL + "/"
				}
				urls = append(urls, i.HomepageURL)
			}
		}
	}

	results := make(chan string, len(urls))
	pageMap := make(map[string]string)

	jobPaths := []string{
		"careers",
		"career",
		"jobs",
		"job",
	}
	for _, url := range urls {
		go func() {
			if strings.Contains(url, "github") {
				results <- fmt.Sprintf("No github %s", url)
				return
			}
			for _, p := range jobPaths {
				status := checkCareersPage(url+p)
				if status == "200 OK" {
					results <- url+p
					return
				}
			}
			results <- ""
		}()

	}
	for range urls {
		page := <-results
		if page != "" {
			re := regexp.MustCompile(`(?:https?://)?(?:www\.)?([^/]+)`)
			match := re.FindStringSubmatch(page)
			if len(match) > 1 {
				pageMap[match[1]] = page
			} else {
				pageMap[page] = page
			}
		}
	}

	    // m is a map[string]interface.
    // loop over keys and values in the map.
	fmt.Println("Found Career pages")
	fmt.Println("| project | career page |")
	fmt.Println("| --- | --- |")
    for k, v := range pageMap {
        fmt.Println("| ", k, " | ", v, " |")
    }
}

func checkCareersPage(url string) string {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println(err)
		return fmt.Sprintf("%s %s", url, err)
	}
	defer resp.Body.Close()
	return resp.Status
}
