all:
	go build -buildmode=c-shared -o out_stackdriver.so .

fast:
	go build out_stackdriver.go

build-linux:
	docker run -v "${PWD}":/go/src/myapp -w /go/src/myapp -it --rm golang:1.8 sh -c 'curl https://glide.sh/get | sh && glide install && make'

build-image:
	docker build -t asia.gcr.io/apstndb-sandbox/fluent-bit-stackdriver:0.1.0 .
push-image:
	docker push asia.gcr.io/apstndb-sandbox/fluent-bit-stackdriver:0.1.0
clean:
	rm -rf *.so *.h *~
