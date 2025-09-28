package models

type RouteDistanceRequest struct {
    Path []string `json:"path"`
}

type CountByStopsRequest struct {
    From     string `json:"from"`
    To       string `json:"to"`
    MinStops int    `json:"minStops"`
    MaxStops int    `json:"maxStops"`
}

type CountByDistanceRequest struct {
    From        string `json:"from"`
    To          string `json:"to"`
    MaxDistance int    `json:"maxDistance"`
}

type RouteSearchConstraints struct {
	MaxStops      int  `json:"maxStops,omitempty"`
	MaxDistance   int  `json:"maxDistance,omitempty"`
	DistinctNodes bool `json:"distinctNodes,omitempty"`
}

type RouteSearchRequest struct {
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	Constraints RouteSearchConstraints `json:"constraints"`
	Limit       int                    `json:"limit"`
}

type RouteSearchResponse struct {
	Routes []struct {
		Path     []string `json:"path"`
		Distance int      `json:"distance"`
	} `json:"routes"`
}