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
	rm -f apitest
	rm -f repotest
	rm -f modeltest
	go mod tidy

.PHONY: test
test: | depcheck modeltest repotest apitest

.PHONY: depcheck
depcheck: 
	go get -u
	#Use go tidy before release testing

#TODO: Change most of these to use stamps naming / hidden files
apitest: api api/*.go
	cd api && go test -cover -coverprofile=coverage.out
	cd api && go vet
	cd api && go tool cover -html=coverage.out
	@touch $@

repotest: repository repository/*.go
	cd repository && go test -cover -coverprofile=coverage.out
	cd repository && go vet
	cd repository && go tool cover -html=coverage.out
	@touch $@

modeltest: model model/*.go
	cd model && go test -cover -coverprofile=coverage.out
	cd model && go vet
	cd model && go tool cover -html=coverage.out
	@touch $@

.PHONY: deploy
deploy: clean build
	sls deploy --verbose
