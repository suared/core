.PHONY: build
build:
	go build
	#env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/process process/main.go
	#chmod +x bin/process
	#zip -j bin/process.zip bin/process
	#rm bin/process 

.PHONY: commitcheck
commitcheck: clean test

.PHONY: clean
clean:
	rm -f infratest
	rm -f apitest
	rm -f ziptest
	rm -f testtest
	rm -f securitytest
	go mod tidy

.PHONY: test
test: | depcheck infratest apitest ziptest testtest securitytest

.PHONY: depcheck
depcheck: 
	#Uncomment before check in, while offline
	#go get -u #Not required with go mod

#TODO: Change most of these to use stamps convention/ hidden files
infratest: infra infra/*.go
	cd infra && go test -cover -coverprofile=coverage.out
	cd infra && go vet
	cd infra && go tool cover -html=coverage.out
	@touch $@

apitest: api api/*.go
	cd api && go test -cover -coverprofile=coverage.out
	cd api && go vet
	cd api && go tool cover -html=coverage.out
	@touch $@

testtest: test test/*.go
	cd test && go test -cover -coverprofile=coverage.out
	cd test && go vet
	cd test && go tool cover -html=coverage.out
	@touch $@

ziptest: ziptools ziptools/*.go
	cd ziptools && go test -cover -coverprofile=coverage.out
	cd ziptools && go vet
	cd ziptools && go tool cover -html=coverage.out
	@touch $@

securitytest: security security/*.go
	cd security && go test -cover -coverprofile=coverage.out
	cd security && go vet
	cd security && go tool cover -html=coverage.out
	@touch $@

.PHONY: deploy
deploy: clean build
	#sls deploy --verbose
