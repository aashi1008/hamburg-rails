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
