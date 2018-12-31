.PHONY: build
build:
	go build
	#env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/process process/main.go
	#chmod +x bin/process
	#zip -j bin/process.zip bin/process
	#rm bin/process 

.PHONY: clean
clean:
	rm -f infratest

.PHONY: test
test: | depcheck infratest infraapitest ziptest

#No longer called, this is now handled by modules
.PHONY: depcheck
depcheck: 
	go get -u

#TODO: Change most of these to use stamps convention/ hidden files
infratest: infra infra/*.go
	cd infra && go test -cover -coverprofile=coverage.out
	cd infra && go vet
	cd infra && go tool cover -html=coverage.out
	@touch $@

infraapitest: infra/api infra/api/*.go
	cd infra/api && go test -cover -coverprofile=coverage.out
	cd infra/api && go vet
	cd infra/api && go tool cover -html=coverage.out
	@touch $@

ziptest: ziptools ziptools/*.go
	cd ziptools && go test -cover -coverprofile=coverage.out
	cd ziptools && go vet
	cd ziptools && go tool cover -html=coverage.out
	@touch $@

.PHONY: deploy
deploy: clean build
	#sls deploy --verbose
