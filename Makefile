run:
	go run cmd/obc-watcher/main.go

build:
	go build  -o dist/ cmd/obc-watcher/main.go

docker-build:
	docker build --platform linux/amd64 -t quay.io/mohamedf0/obc-watcher:v1 .

docker-build-s390x:
	docker build --platform linux/s390x -f Dockerfile.s390x -t quay.io/mohamedf0/obc-watcher:v1-s390x  .

docker-push:
	docker push quay.io/mohamedf0/obc-watcher:v1

docker-push-s390x:
	docker push quay.io/mohamedf0/obc-watcher:v1-s390x

deploy:
	if [ "$(NAMESPACE)" = "" ]; then \
		echo "ERROR: MISSING NAMESPACE"; exit 1; \
	fi;

	kubectl apply -f manifests/clusterrole.yaml
	sed 's/@NAMESPACE/$(NAMESPACE)/' manifests/serviceaccount.yaml | kubectl apply -f -
	sed 's/@NAMESPACE/$(NAMESPACE)/' manifests/clusterrolebinding.yaml | kubectl apply -f -
	sed 's/@NAMESPACE/$(NAMESPACE)/' manifests/deployment.yaml | kubectl apply -f -

deploy-s390x:
	kubectl apply -f manifests/clusterrole.yaml
	sed 's/@NAMESPACE/$(NAMESPACE)/' manifests/serviceaccount.yaml | kubectl apply -f -
	sed 's/@NAMESPACE/$(NAMESPACE)/' manifests/clusterrolebinding.yaml | kubectl apply -f -
	cat manifests/deployment.yaml | sed 's/@NAMESPACE/$(NAMESPACE)/'| sed 's/obc-watcher:v1/obc-watcher:v1-s390x/' | kubectl apply -f -