package server

import (
	"embed"
)

//go:embed swagger_files/*
var swaggerContent embed.FS
