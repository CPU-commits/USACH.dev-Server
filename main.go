package main

import (
	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/jobs"
	"github.com/CPU-commits/USACH.dev-Server/server"
)

func main() {
	// Init jobs
	jobs.Init()
	// Register custom validators
	forms.Init()
	// Init server
	server.Init()
}
