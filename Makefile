.PHONY: test cover

test:
	go test ./... -v

cover:
	go test ./... -coverprofile=profiles/cover.out -coverpkg=./internal/...
	go tool cover -func=profiles/cover.out

cover-html:
	go test ./... -coverprofile=profiles/cover.out -coverpkg=./internal/...
	go tool cover -html=profiles/cover.out
