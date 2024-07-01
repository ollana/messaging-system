SHELL := /bin/bash

unit_test:
	cd server && go test -race -timeout=120s -v ./...

build:
	cd server && go build -o ../app

build-image:
	cd server && DOCKER_DEFAULT_PLATFORM="linux/amd64" docker build -t ${AWS_ACCOUNT_ID}.dkr.ecr.us-west-2.amazonaws.com/messaging-system-app:1.0.15 .

push-image: build-image
	# login to docker
	aws ecr get-login-password --region us-west-2 --profile pulumi  | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.us-west-2.amazonaws.com
	# create ecr repository if doesn't exist
	aws ecr create-repository --repository-name messaging-system-app --region us-west-2 --profile pulumi || echo "Repository already exists"
	docker push ${AWS_ACCOUNT_ID}.dkr.ecr.us-west-2.amazonaws.com/messaging-system-app:1.0.15

deploy: push-image
	pulumi up --yes

destroy:
	pulumi destroy --yes
