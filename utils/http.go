package utils

import (
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/PChaparro/go-youtube-scraper/interfaces"
)

// MakeRequestAndGetData Makes an http GET request, obtain the data with regular expressions,
// sanitize it and return a Video interface containing the video url, thumbnail url, title,
// description and tags.
func MakeRequestAndGetData(url string, sharedSlice *interfaces.ConcurrentSlice, wg *sync.WaitGroup, c chan int) {
	defer wg.Done() // Reduce wg

	// ### Get plain html
	resp, err := http.Get(url)

	if err != nil {
		log.Fatal("Unable to make get request")
	}

	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)
	html := string(bytes)

	// ### Parse plain html with regular expressions
	titleRegExp := regexp.MustCompile(`<meta name="title"[^>]+content="(.*?)"`)
	descriptionRegExp := regexp.MustCompile(`"shortDescription":"(.*?)"`)
	tagsRegExp := regexp.MustCompile(`<meta name="keywords"[^>]+content="(.*?)"`)
	thumbnailRegExp := regexp.MustCompile(`<link rel="image_src"[^>]+href="(.*?)"`)

	titleEvalResults := titleRegExp.FindStringSubmatch(html)
	descriptionEvalResults := descriptionRegExp.FindStringSubmatch(html)
	tagsEvalResults := tagsRegExp.FindStringSubmatch(html)
	thumbnailEvalResults := thumbnailRegExp.FindStringSubmatch(html)

	var (
		title       string
		description string
		tagsString  string
		thumbnail   string
	)

	if len(titleEvalResults) > 1 {
		title = titleEvalResults[1]
	}

	if len(descriptionEvalResults) > 1 {
		description = descriptionEvalResults[1]
	}

	if len(tagsEvalResults) > 1 {
		tagsString = tagsEvalResults[1]
	}

	if len(thumbnailEvalResults) > 1 {
		thumbnail = thumbnailEvalResults[1]
	}

	// ### Basic sanitization on title and description
	title = Sanitize(title)
	description = Sanitize(description)

	// Remove trailing spaces on tags
	tagsArray := strings.Split(tagsString, ",")
	newTagsArray := []string{}

	for _, tag := range tagsArray {
		newTag := strings.Trim(tag, " ")

		if newTag != "" {
			newTagsArray = append(newTagsArray, newTag)
		}
	}

	// ### Creates and add the Video interface to the shared array
	video := interfaces.Video{
		Url:         url,
		Title:       title,
		Description: description,
		Thumbnail:   thumbnail,
		Tags:        newTagsArray,
	}

	sharedSlice.Lock()
	defer sharedSlice.Unlock()
	sharedSlice.Items = append(sharedSlice.Items, video)

	<-c // Reduce channel
}
