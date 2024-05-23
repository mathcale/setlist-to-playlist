package setlistfm

import "fmt"

type Artist struct {
	MBID           string `json:"mbid"`
	Name           string `json:"name"`
	SortName       string `json:"sortName"`
	Disambiguation string `json:"disambiguation"`
	URL            string `json:"url"`
}

type Coords struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

type Country struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type City struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	State     string  `json:"state"`
	StateCode string  `json:"stateCode"`
	Coords    Coords  `json:"coords"`
	Country   Country `json:"country"`
}

type Venue struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	City City   `json:"city"`
	URL  string `json:"url"`
}

type Tour struct {
	Name string `json:"name"`
}

type Song struct {
	Name string `json:"name"`
}

type Songs struct {
	Song   []Song `json:"song"`
	Encore int    `json:"encore,omitempty"`
}

type Sets struct {
	Set []Songs `json:"set"`
}

type Set struct {
	ID          string `json:"id"`
	VersionID   string `json:"versionId"`
	EventDate   string `json:"eventDate"`
	LastUpdated string `json:"lastUpdated"`
	Artist      Artist `json:"artist"`
	Venue       Venue  `json:"venue"`
	Tour        Tour   `json:"tour"`
	Sets        Sets   `json:"sets"`
	URL         string `json:"url"`
}

func (s *Set) Title() string {
	return fmt.Sprintf("%s %s @ %s, %s - %s", s.Artist.Name, s.Tour.Name, s.Venue.Name, s.Venue.City.Name, s.Venue.City.Country.Name)
}

func (s *Set) Songs() []string {
	var songs []string

	for _, set := range s.Sets.Set {
		for _, song := range set.Song {
			songs = append(songs, song.Name)
		}
	}

	return songs
}

func (s *Set) ArtistName() string {
	return s.Artist.Name
}
