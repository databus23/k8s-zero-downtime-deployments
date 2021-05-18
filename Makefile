IMAGE := keppel.eu-de-1.cloud.sap/ccloud/zero-downtime-test-webserver

VERSION := 0.1.4

build:
	docker build -t $(IMAGE):$(VERSION) .

push:
	docker push $(IMAGE):$(VERSION)

load:
	cat urls.txt  | vegeta attack  --duration=30s | vegeta report --every=1s
show:
	@kustomize build deployment/ | bat -l yaml --paging never
deploy:
	kustomize build deployment/ | kubectl apply -f -
