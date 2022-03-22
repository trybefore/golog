
#might change later
CONFIG_PATH=${HOME}/.golog/


.PHONY: compile
compile:
	protoc api/*.proto --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \

.PHONY: test
test:
	go test -race ./...

.PHONY: bench
bench:
	go test -bench=. ./...

.PHONY: gencerts
gencerts:
	cfssl gencert -initca test/ca-csr.json | cfssljson -bare ca
	cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=test/ca-config.json -profile=server test/server-csr.json | cfssljson -bare server 
	cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=test/ca-config.json -profile=client test/client-csr.json | cfssljson -bare client

	mv *.pem *.csr ${CONFIG_PATH}