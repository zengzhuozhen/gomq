.PHONY: dockerPrepare build run benchmarkTest

#构建docker镜像
dockerPrepare:
	docker build -t gomq/gomq:v1 .

#生成二进制运行文件
build:
	go build -o gomq server/main.go
	go build -o gomqctl cmd/main.go

benchmarkTest:
	ab -n 1000 -c 10 -T 'application/x-www-form-urlencode' -p benchmark/input-level-1.txt  http://localhost:8000/pub > benchmark/output-level-1.txt
	ab -n 1000 -c 10 -T 'application/x-www-form-urlencode' -p benchmark/input-level-2.txt http://localhost:8000/pub > benchmark/output-level-2.txt
	ab -n 1000 -c 10 -T 'application/x-www-form-urlencode' -p benchmark/input-level-3.txt  http://localhost:8000/pub > benchmark/output-level-3.txt



#运行
run:
	go run -v ./server/main.go
