package types

import "time"

type Photo struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type ImageInfo struct {
	URL      string
	Key      string
	Size     int64
	Modified time.Time
}
