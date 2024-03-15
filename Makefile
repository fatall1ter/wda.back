APPNAME = wda.back
# h - help
h help:
	@echo "h help 	- this help"
	@echo "build 	- build and the app"
	@echo "run 	- run the app"
	@echo "clean 	- clean app trash"
	@echo "dev 	- generate docs and run"
	@echo "test 	- run all tests"
	@echo "docker 	- run docker image build"
.PHONY: h

# build - build the app
build:
#	export GOPRIVATE=*.watcom.ru
	go build -o $(APPNAME)
.PHONY: build

# run - build and run the app
run: build
	./$(APPNAME) -p=8088  -level=debug
.PHONY: run

clean:
	rm ./$(APPNAME)
.PHONY: clean

# dev - generate docs and run
dev: run
.PHONY: dev

# test - run all tests
test:
	go test ./...
.PHONY: test

# docker build
docker:
	docker build . -t hub.watcom.ru/$(APPNAME)
.PHONY: docker