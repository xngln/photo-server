// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"github.com/99designs/gqlgen/graphql"
)

type Image struct {
	ID           string  `json:"_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	ThumbnailURL string  `json:"thumbnail_url"`
	FullsizeURL  string  `json:"fullsize_url"`
}

type NewImage struct {
	Name  string         `json:"name"`
	Price float64        `json:"price"`
	File  graphql.Upload `json:"file"`
}
