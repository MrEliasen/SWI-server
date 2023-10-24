package game

import (
	"fmt"

	"github.com/mreliasen/swi-server/internal/responses"
)

type Coordinates struct {
	North   int    `json:"north"`
	East    int    `json:"east"`
	City    string `json:"city,omitempty"`
	POI     string `json:"poi,omitempty"`
	POIType uint8  `json:"poitype,omitempty"`
}

func (c *Coordinates) toResponse() *responses.Coordinate {
	return &responses.Coordinate{
		North:   int32(c.North),
		East:    int32(c.East),
		City:    c.City,
		Poi:     c.POI,
		Poitype: int32(c.POIType),
	}
}

func (c *Coordinates) toString() string {
	return fmt.Sprintf("N%d-E%d", c.North, c.East)
}

func (c *Coordinates) SameAs(c2 *Coordinates) bool {
	if c2 == nil {
		return false
	}

	return c.East == c2.East && c.North == c2.North
}
