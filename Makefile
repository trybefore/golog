
#might change later
CONFIG_PATH=${HOME}/.golog/

${CONFIG_PATH}/policy.csv:
	cp test/policy.csv ${CONFIG_PATH}/policy.csv

${CONFIG_PATH}/model.conf:
	cp test/model.conf ${CONFIG_PATH}/model.conf


.PHONY: compile
compile:
	protoc api/*.proto --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \

.PHONY: test
test: ${CONFIG_PATH}/policy.csv ${CONFIG_PATH}/model.conf
	go test ./...

.PHONY: bench
bench:
	go test -bench=. ./...

.PHONY: gencerts
gencerts:
	cfssl gencert -initca test/ca-csr.json | cfssljson -bare ca
	cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=test/ca-config.json -profile=server test/server-csr.json | cfssljson -bare server 
	cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=test/ca-config.json -profile=client test/client-csr.json | cfssljson -bare client
	cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=test/ca-config.json -profile=client -cn="root" test/client-csr.json | cfssljson -bare root-client
	cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=test/ca-config.json -profile=client -cn="unauthorized" test/client-csr.json | cfssljson -bare unauthorized-client

	mv *.pem *.csr ${CONFIG_PATH}