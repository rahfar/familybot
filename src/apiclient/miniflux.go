package apiclient

import (
	"fmt"
	"strings"

	miniflux "miniflux.app/client"
)

type MinifluxAPI struct {
	BaseURL string
	ApiKey  string
	SiteURL string
}

func (m *MinifluxAPI) GetLatestNews() (miniflux.Entries, error) {
	client := miniflux.New(m.BaseURL, m.ApiKey)

	// Fetch all feeds.
	feeds, err := client.Feeds()
	if err != nil {
		return miniflux.Entries{}, err
	}
	var myFeed *miniflux.Feed
	for _, f := range feeds {
		if strings.HasPrefix(f.SiteURL, m.SiteURL) {
			myFeed = f
		}
	}
	if myFeed.ID != 0 {
		entries, err := client.Entries(&miniflux.Filter{FeedID: myFeed.ID, Limit: 3, Order: "published_at", Direction: "desc"})
		if err != nil {
			return miniflux.Entries{}, err
		}
		return entries.Entries, nil
	} else {
		return miniflux.Entries{}, fmt.Errorf("no feed")
	}
}
