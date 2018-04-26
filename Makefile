.dockerbuilt: Dockerfile redirector.go
	docker build -t redirector .
	docker tag redirector localhost:5000/redirector
	docker push localhost:5000/redirector
	touch .dockerbuilt

build:	.dockerbuilt

run:	.dockerbuilt
	docker run redirector
