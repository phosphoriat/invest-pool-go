package models

type Problem struct {
	ID        string     `json:"id"`
	RegionID  string     `json:"region_id"`
	Title     string     `json:"title"`
	Solutions []Solution `json:"solutions"`
	Match     *Match     `json:"match"`
	Pool      *Pool      `json:"pool"`
}
