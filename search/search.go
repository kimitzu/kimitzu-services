package search

import (
	"fmt"
	"strings"

	"gitlab.com/kingsland-team-ph/djali/djali-services.git/models"
)

// Find the listings and returns potential matches via supplied keyword
func Find(keyword string, averageRating int64, listings []*models.Listing) []*models.Listing {
	fmt.Println(keyword)
	response := []*models.Listing{}
	for _, listing := range listings {
		if findByKeyword(keyword, listing) && findByAverageRating(averageRating, listing) {
			response = append(response, listing)
		}
	}
	return response
}

func findByKeyword(keyword string, listing *models.Listing) bool {
	keywordLowercase := strings.ToLower(keyword)
	// Probably an initial wildcard search or just browsing via filters
	if keyword == "" {
		return true
	}
	return strings.Contains(strings.ToLower(listing.Listing.Item.Title), keywordLowercase) || strings.Contains(strings.ToLower(listing.Listing.Item.Description), keywordLowercase)
}

func findByAverageRating(averageRating int64, listing *models.Listing) bool {
	if averageRating <= 0 {
		return true
	}
	return listing.AverageRating >= averageRating
}
