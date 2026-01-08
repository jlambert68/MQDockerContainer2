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

Start_DockerCompose_MQClientAndMQHost:
	cd mq-gateway && docker compose up -d

Start_DockerCompose_ForceBuild_MQClientAndMQHost:
	docker  compose up --build -d

Stop_DockerCompose_MQClientAndMQHost:
	cd MqClientAndMqHost && docker compose stop

Logs_MQ_Gateway:
	docker logs -f --tail 100 mq-gateway

Logs_MQ_Host:
	docker logs -f --tail 100 mq-host

docker_Just build_with_debug_MQClientAndMQHost:
	#cd MqClientAndMqHost && docker build --no-cache --progress=plain -t mq-gateway-debug .
	cd mq-gateway && docker compose build --no-cache --progress=plain
#docker compose build --no-cache --progress=plain
#docker compose build --no-cache --pull --progress=plain

docker_list_all_containers:
	docker ps -a

docker_list_running_containers:
#	docker ps
	docker ps -a --filter "status=running"

ListFilderTree:
	tree