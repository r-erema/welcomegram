build:
	rm -rf build && mkdir build && GOOS=linux go build -o build/welcomegram main.go && cp selenium-server-standalone-3.141.59.jar build && cp chromedriver_linux build && cp config.json.dist build && zip -r welcomegram.zip build && rm -rf build
