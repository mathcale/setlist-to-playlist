package music

import (
	"sort"
	"strings"
)

var liveMarkers = []string{"(live", "- live", "live at", "live from", "unplugged", "session"}

// IsLive reports whether the candidate looks like a live/session recording,
// judging by its title or album name.
func (c Candidate) IsLive() bool {
	text := strings.ToLower(c.Title + "\n" + c.Album)
	for _, m := range liveMarkers {
		if strings.Contains(text, m) {
			return true
		}
	}

	return false
}

// tier: lower is better. Order: album/EP > compilation > single > unknown > live.
// Unknown album types are ranked after single (nothing makes them better) but
// before live (they are not known to be live). Live candidates rank after every
// studio one, ordered among themselves by the same album type order.
func (c Candidate) tier() int {
	if c.IsLive() {
		return liveTierOffset + c.albumTier()
	}

	return c.albumTier()
}

const liveTierOffset = 4

func (c Candidate) albumTier() int {
	switch c.AlbumType {
	case AlbumTypeAlbum, AlbumTypeEP:
		return 0
	case AlbumTypeCompilation:
		return 1
	case AlbumTypeSingle:
		return 2
	default:
		return 3
	}
}

// PickBestCandidate drops candidates by other artists and returns the best
// remaining one (studio first, live as last resort; stable within a tier).
func PickBestCandidate(candidates []Candidate, artist string) (Candidate, bool) {
	want := strings.ToLower(strings.TrimSpace(artist))

	var kept []Candidate
	for _, c := range candidates {
		if strings.ToLower(strings.TrimSpace(c.Artist)) == want {
			kept = append(kept, c)
		}
	}

	if len(kept) == 0 {
		return Candidate{}, false
	}

	sort.SliceStable(kept, func(i, j int) bool { return kept[i].tier() < kept[j].tier() })

	return kept[0], true
}
