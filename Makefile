IMAGE := keppel.eu-de-1.cloud.sap/ccloud/zero-downtime-test-webserver

VERSION := 0.1.5

build:
	docker build -t $(IMAGE):$(VERSION) .

push:
	docker push $(IMAGE):$(VERSION)

load:
	cat urls.txt  | vegeta attack  --timeout=40s --duration=30s | vegeta report --every=1s
show:
	@kustomize build | bat -l yaml --paging never
deploy:
	kustomize build | kubectl apply -f -
