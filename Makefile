compileProto_go:
	@echo "Compile proto file..."

 # generate the messages
	protoc --go_out=mq-gateway/api/proto mq-gateway/api/proto/*.proto

ListDockerContainersWithStatus:
	docker ps -a

RunDockerCompose_MQClientAndMQHost:
	cd MqClientAndMqHost && docker compose up -d

StopDockerCompose_MQClientAndMQHost:
	cd MqClientAndMqHost && docker compose stop