run:
	go run cmd/obc-watcher/main.go

build:
	go build  -o dist/ cmd/obc-watcher/main.go

docker-build:
	docker build -t quay.io/mohamedf0/obc-watcher:v1 .

docker-build-s390x:
	docker build --platform linux/s390x -f Dockerfile.s390x -t quay.io/mohamedf0/obc-watcher:v1-s390x  .

docker-push:
	docker push quay.io/mohamedf0/obc-watcher:v1

docker-push-s390x:
	docker push quay.io/mohamedf0/obc-watcher:v1-s390x

deploy:
	kubectl apply -f manifests/namespace.yaml
	kubectl apply -f manifests/clusterrole.yaml
	kubectl apply -f manifests/clusterrolebinding.yaml
	kubectl apply -f manifests/serviceaccount.yaml
	kubectl apply -f manifests/deployment.yaml

deploy-s390x:
	kubectl apply -f manifests/namespace.yaml
	kubectl apply -f manifests/clusterrole.yaml
	kubectl apply -f manifests/clusterrolebinding.yaml
	kubectl apply -f manifests/serviceaccount.yaml
	sed 's/obc-watcher:v1/obc-watcher:v1-s390x/' manifests/deployment.yaml | kubectl apply -f -