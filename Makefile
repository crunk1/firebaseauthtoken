build:
	@go build -o firebaseauthtoken cmd/firebaseauthtoken.go

install: build
	@cp firebaseauthtoken ${GOPATH}/bin/firebaseauthtoken 
