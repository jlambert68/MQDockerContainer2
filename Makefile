compileProto_go:
	@echo "Compile proto file..."

 # generate the messages
#	protoc --go_out=mq-gateway/api/proto mq-gateway/api/proto/*.proto

 # generate the messages
# protoc --go_out="$GO_GEN_PATH" -I "$dependecies" "$proto"
	cd mq-gateway/api/proto && protoc --go_out=. mq.proto

# generate the services
# protoc --go-grpc_out="$GO_GEN_PATH" -I "$dependecies" "$proto"
	cd mq-gateway/api/proto && protoc --go-grpc_out=. mq.proto

ListDockerContainersWithStatus:
	docker ps -a

RunDockerCompose_MQClientAndMQHost:
	cd MqClientAndMqHost && docker compose up -d

StopDockerCompose_MQClientAndMQHost:
	cd MqClientAndMqHost && docker compose stop