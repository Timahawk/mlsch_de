BINARY_NAME=mlsch_de

hello:
	echo "Hello"

build:
	#	go build -o bin/mlsch_de .
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY_NAME}-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o bin/${BINARY_NAME}-windows-amd64.exe .

deploy:
	export PATH=$PATH:/usr/local/go/bin
	echo "Compiling for OS and Platform"
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY_NAME} .
	echo "Calling Ansilbe Playbook"
	ansible-playbook ./scripts/deploy_to_Lightsail.yml


run:
	go run .

clean:
	go clean
	find ./bin/ -name ${BINARY_NAME}* -delete