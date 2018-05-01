package commute

import (
	"testing"
	"math"
)

func TestDist(t *testing.T) {
	cases := []struct {
		p1, p2 Point 
		dist float64
	}{
		{ Point{Lat: 38.89768, Lon: -77.03653}, Point{Lat: 38.89768, Lon: -77.03653}, 0},
		{ Point{Lat: 12.884733, Lon: 77.551541}, Point{Lat: 12.918230, Lon: 77.573472}, 4418},
		{ Point{Lat: 12.884733, Lon: 77.551541}, Point{Lat: 12.975995, Lon: 77.572847}, 10408},
	}
	for idx, c := range cases {
		got := DistanceBetwnPts(c.p1, c.p2)
		if math.Abs(got - c.dist) > 1 {
			t.Errorf("distBetwnPoints(%q) == %q, want %q", idx, got, c.dist)
		}
	}

}