package youtubescraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/PChaparro/go-youtube-scraper/interfaces"
	"github.com/PChaparro/go-youtube-scraper/utils"
	"github.com/go-rod/rod" // Web driver
)

// GetVideosUrlFromApi Make a GET request to youtube API to get videos related to the
// search criteria and return the urls
func GetVideosUrlFromApi(youtubeApiKey, searchCriteria string, size int) ([]string, error) {
	if size <= 0 {
		log.Fatalf(fmt.Sprintf("Desired array length must be greater than zero. Got: %d", size))
	}

	// Slice to store the videos urls
	urls := []string{}
	uniques := make(map[string]uint)

	// Replace blank spaces
	searchCriteria = strings.ReplaceAll(searchCriteria, " ", "%20")
	pageKey := "" // Initially empty to go the first page

	// Repeat until wanted size is reached
	for len(urls) != size {
		// It's not necessary to limit the size for the request because
		// youtube API does it automatically
		uri := fmt.Sprintf("https://youtube.googleapis.com/youtube/v3/search?maxResults=%d&q=%s&pageToken=%s&type=video&key=%s", size, searchCriteria, pageKey, youtubeApiKey)

		// Http request
		resp, err := http.Get(uri)

		if err != nil {
			return []string{}, err
		}

		defer resp.Body.Close()

		// Parse request into struct
		reply := interfaces.SearchVideosReply{}
		bytes, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return []string{}, err
		}

		err = json.Unmarshal(bytes, &reply)
		if err != nil {
			return []string{}, err
		}

		// Lop for each video and get it's url
		for _, item := range reply.Items {
			url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.IdGroup.VideoId)

			// Don't allow repeated videos
			if uniques[url] != 0 {
				continue
			}

			urls = append(urls, url)

			if len(urls) == size {
				break
			}
		}

		// Replace the next page to request more different videos
		pageKey = reply.NextPageToken
	}

	return urls, nil
}

// GetVideosUrlFromSite Open a web-browser headless instance and scroll the page until
// obtain the desired amount of videos and return it's urls.
func GetVideosUrlFromSite(searchCriteria string, size int) ([]string, error) {
	if size <= 0 {
		log.Fatal(fmt.Sprintf("Desired array length must be greater than zero. Got: %d", size))
	}

	// Replace blank spaces and create the url
	searchCriteria = strings.ReplaceAll(searchCriteria, " ", "+")
	target := fmt.Sprintf("https://www.youtube.com/results?search_query=%s", searchCriteria)

	// Slice to store the videos urls
	urls := []string{}

	// Create the headless web instance
	browser := rod.New().MustConnect()
	defer browser.MustClose()
	page := browser.MustPage(target).MustWaitLoad()
	defer page.MustClose()

	// Count current video urls items on page
	currentElements := page.MustElements("#video-title")

	// Scroll until get the desired amount of videos plus 10 more elements to get an
	// error margin
	for len(currentElements) < (size + 10) {
		// TODO: This doesn't await the scroll to be completed
		page.Eval("window.scroll({top: 9999999999})")

		// Count again
		currentElements = page.MustElements("#video-title")
	}

	// Get the links
	for _, element := range currentElements {
		val := element.MustAttribute("href")

		if val != nil {
			// Each *val has the form: /watch?v=ysEN5RaKOlA, so, we have to convert to an absolute path
			urls = append(urls, fmt.Sprintf("https://youtube.com%s", *val))
		}

		if len(urls) == size {
			// Break when the desired amount was obtained
			break
		}
	}

	return urls, nil
}

// GetVideosData Get the videos urls with the desired method (With and without using the youtube api),
// and then, obtain the title, description, thumbnail, tags and url for each video.
func GetVideosData(youtubeApiKey, criteria string, size, concurrencyLimit int, useYoutubeApi bool) (interfaces.Videos, error) {
	// Simple validations
	if useYoutubeApi && (youtubeApiKey == "") {
		log.Fatal(fmt.Sprintf("You have to provide a youtube api key if the useYoutubeApi argument is defined as true"))
	}

	if size < 0 {
		log.Fatalf(fmt.Sprintf("Desired array length must be greater than zero. Got: %d", size))
	}

	if concurrencyLimit < 0 {
		log.Fatalf(fmt.Sprintf("Concurrency limit must be an integer greater than zero. Got: %d", size))
	}

	// Channel to define concurrency limit
	semaphore := make(chan int, concurrencyLimit)
	var wg sync.WaitGroup

	// Get urls with the desired method
	urls := []string{}

	if useYoutubeApi {
		res, err := GetVideosUrlFromApi(youtubeApiKey, criteria, size)

		if err != nil {
			return interfaces.Videos{}, nil
		}

		urls = res
	} else {
		res, err := GetVideosUrlFromSite(criteria, size)

		if err != nil {
			return interfaces.Videos{}, nil
		}

		urls = res
	}

	// Get data for each video and save on the shared slice
	sharedSlice := interfaces.ConcurrentSlice{}

	for _, url := range urls {
		semaphore <- 1
		wg.Add(1)
		go utils.MakeRequestAndGetData(url, &sharedSlice, &wg, semaphore)
	}

	wg.Wait()

	// Get data for each url
	return interfaces.Videos{
		Videos: sharedSlice.Items,
	}, nil

}
