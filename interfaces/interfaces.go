package interfaces

import (
	"sync"
)

// Youtube API response interfaces
type VideoId struct {
	VideoId string `json:"videoId"`
}

type VideoItem struct {
	IdGroup VideoId `json:"id"`
}

type SearchVideosReply struct {
	NextPageToken string      `json:"nextPageToken"`
	Items         []VideoItem `json:"items"`
}

// Package responses interfaces
type Video struct {
	Url         string   `json:"url"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Thumbnail   string   `json:"thumbnail"`
	Tags        []string `json:"tags"`
}

type Videos struct {
	Videos []Video `json:"videos"`
}

// Helpers interfaces
type ConcurrentSlice struct {
	sync.RWMutex
	Items []Video
}
