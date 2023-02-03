BINARY_NAME=mlsch_de

# export PATH=$PATH:/usr/local/go/bin

hello:
	echo "Hello"

build:
	#	go build -o bin/mlsch_de .
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY_NAME}-linux-amd64 .
	GOOS=linux GOARCH=arm GOARM=7 go build -o bin/${BINARY_NAME}-linux-arm7 .
	GOOS=windows GOARCH=amd64 go build -o bin/${BINARY_NAME}-windows-amd64.exe .

deploy:
	echo "Deploying to Lightsail instance."
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY_NAME}-linux-amd64 .
	echo "Calling Ansilbe Playbook"
	ansible-playbook ./scripts/deploy_to_Lightsail.yml

run:
	go run .

clean:
	go clean
	find ./bin/ -name ${BINARY_NAME}* -delete