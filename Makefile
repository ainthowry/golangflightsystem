install:
	go mod download
	
serve:
	go run cmd/main.go

tidy:
	go mod tidy