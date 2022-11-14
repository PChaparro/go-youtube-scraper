package tests

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	youtubescraper "github.com/PChaparro/go-youtube-scraper"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

// Check the youtube api key was provided
func TestEnvironment(t *testing.T) {
	c := require.New(t)

	err := godotenv.Load()
	c.NoError(err)

	key := os.Getenv("YOUTUBE_API_KEY")
	c.NotEqualf("", key, "Youtube api key must be provided")
}

func TestGetVideosUrlSuccess(t *testing.T) {
	c := require.New(t)

	key := os.Getenv("YOUTUBE_API_KEY")

	expectedLength := 100
	urls, err := youtubescraper.GetVideosUrlFromApi(key, "Web Development", expectedLength)

	// 1. Check the array has the desired length
	c.NoError(err)
	c.Equalf(expectedLength, len(urls), fmt.Sprintf("Exptected %d videos but got %d", expectedLength, len(urls)))

	// 2. Check all the texts are valid youtube videos urls
	youtubeUrlRegexp := regexp.MustCompile(`(https?://)?(www\.)?(youtube|youtu|youtube-nocookie)\.(com|be)/(watch\?v=|embed/|v/|.+\?v=)?(?P<id>[A-Za-z0-9\-=_]{11})`)

	repetitions := make(map[string]int) // For next test

	for _, url := range urls {
		repetitions[url]++
		c.Equalf(true, youtubeUrlRegexp.Match([]byte(url)), fmt.Sprintf("Expected video url to match with youtube url regexp"))
	}

	// 3. Check all the links are unique
	for url := range repetitions {
		c.Equalf(1, repetitions[url], fmt.Sprintf("Expected videos not to be repeated. %s url is repeated %d times", url, repetitions[url]))
	}

}

func TestGetVideosDataSuccess(t *testing.T) {
	c := require.New(t)
	key := os.Getenv("YOUTUBE_API_KEY")

	expectedLength := 100
	concurrencyLimit := 16

	videos, err := youtubescraper.GetVideosData(key, "Web development", expectedLength, concurrencyLimit, true)

	// 1. Check the array has the desired length
	c.NoError(err)
	c.Equalf(expectedLength, len(videos.Videos), fmt.Sprintf("Expected %d videos but got %d", expectedLength, len(videos.Videos)))

	// 2. Check "important" fields
	youtubeUrlRegexp := regexp.MustCompile(`(https?://)?(www\.)?(youtube|youtu|youtube-nocookie)\.(com|be)/(watch\?v=|embed/|v/|.+\?v=)?(?P<id>[A-Za-z0-9\-=_]{11})`)

	for _, video := range videos.Videos {
		c.NotEqualf("", video.Title, fmt.Sprintf("Expected video title not to be empty"))
		c.NotEqualf("", video.Thumbnail, fmt.Sprintf("Expected video thumbnail not to be empty"))
		c.Equalf(true, youtubeUrlRegexp.Match([]byte(video.Url)), fmt.Sprintf("Expected video url to match with youtube url regexp"))
	}
}
