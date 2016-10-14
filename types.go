package main

import "github.com/drone/drone-go/drone"

// Params are the parameters that the GitHub Release plugin can parse.
type Params struct {
	BaseURL  string            `json:"base_url"`
	User     string            `json:"user"`
	Password string            `json:"password"`
	Files    drone.StringSlice `json:"files"`
}
