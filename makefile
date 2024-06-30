SHELL := /bin/bash

unit_test:
	cd server && go test -race -timeout=120s -v ./...

build:
	cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../app

#package: build
#	# create s3 bucket if not exists
#	aws s3api create-bucket --bucket messaging-system-app --region us-west-2 --create-bucket-configuration LocationConstraint=us-west-2 --profile pulumi || echo "Bucket already exists"
#	# upload app to s3 bucket
#	cd server && aws s3api  put-object --body app  --bucket messaging-system-app --key app --profile pulumi

build-image:
	cd server && DOCKER_DEFAULT_PLATFORM="linux/amd64" docker build -t 339713044858.dkr.ecr.us-west-2.amazonaws.com/messaging-system-app:0.0.13 .


# todo: account id cant be cnstant
push-image: build-image
	# create ecr repository if not exists
	aws ecr create-repository --repository-name messaging-system-app --region us-west-2 --profile pulumi || echo "Repository already exists"
	docker push 339713044858.dkr.ecr.us-west-2.amazonaws.com/messaging-system-app:0.0.13

deploy: push-image
	pulumi up --yes

destroy:
	pulumi destroy --yes
