module github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service

go 1.22

require (
	github.com/alexey-y-a/go-taskqueue-microservices/libs/logger v0.0.0
	github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel v0.0.0-00010101000000-000000000000
	github.com/rs/zerolog v1.34.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	golang.org/x/sys v0.12.0 // indirect
)

replace github.com/alexey-y-a/go-taskqueue-microservices/libs/logger => ../../libs/logger

replace github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel => ../../libs/taskmodel
