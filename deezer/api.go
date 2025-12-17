package deezer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Track struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Artist   string
	Album    string
	Duration int    `json:"duration"`
	Preview  string `json:"preview"`
}

type Artist struct {
	Name string `json:"name"`
}

type Album struct {
	Title string `json:"title"`
}

type TrackResponse struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Duration int    `json:"duration"`
	Preview  string `json:"preview"`
	Artist   Artist `json:"artist"`
	Album    Album  `json:"album"`
}

type SearchResponse struct {
	Data []TrackResponse `json:"data"`
}

func SearchTracks(query string) ([]Track, error) {
	url := fmt.Sprintf("https://api.deezer.com/search?q=%s", query)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, err
	}

	var tracks []Track
	for _, t := range searchResp.Data {
		track := Track{
			ID:       t.ID,
			Title:    t.Title,
			Artist:   t.Artist.Name,
			Album:    t.Album.Title,
			Duration: t.Duration,
			Preview:  t.Preview,
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func GetTracksByGenre(genre string) ([]Track, error) {
	genreMap := map[string]string{
		"Rock": "rock",
		"Rap":  "rap",
		"Pop":  "pop",
	}

	searchTerm, ok := genreMap[genre]
	if !ok {
		return nil, errors.New("genre invalide")
	}

	return SearchTracks(searchTerm)
}

func GetRandomTrack(tracks []Track) (*Track, error) {
	if len(tracks) == 0 {
		return nil, errors.New("aucune musique trouvee")
	}

	randomIndex := 0
	if len(tracks) > 1 {
		randomIndex = int(time.Now().UnixNano()) % len(tracks)
	}

	return &tracks[randomIndex], nil
}