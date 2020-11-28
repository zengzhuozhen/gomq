.PHONY: dockerPrepare build run

#构建docker镜像
dockerPrepare:
	docker build -t gomq/gomq:v1 .

#生成二进制运行文件
build:
	go build -o gomq server/main.go
	go build -o gomqctl cmd/main.go


#运行
run:
	go run -v ./server/main.go
