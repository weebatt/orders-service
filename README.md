### Install
First of all you need to do:
```bash
git clone https://gitlab.crja72.ru/golang/2025/spring/course/students/310499-itmobatareyka-course-1343.git
```

### Start
After that you had to opportunities to start the server\
1) Use has already written Makefile:
```bash
make
```

2) Simply in terminal 
```aiignore
go run main/main.go
```

If you'll have some problems don't forget to generate proto  

protoc __with__ grpc-gateway:
```bash
protoc --go_out=./pkg/api/test --go_opt=paths=source_relative \
       --go-grpc_out=./pkg/api/test  --go-grpc_opt=paths=source_relative \
       --grpc-gateway_out=./pkg/api/test --grpc-gateway_opt=paths=source_relative \
       ./proto/api/*proto
```

Adding simple __docs__ of endpoints with swagger:
```bash
protoc -I . --openapiv2_out ./pkg/api/docs ./proto/api/*proto
```

### Description of env variables:
In our project we already have three env var. Their location is ~/config/config.yam. You need to create it yourself.

1) GRPC_PORT - contains with gRPC server port
2) HTTP_PORT - contains with HTTP server port
3) GRPC_SERVER_ENDPOINT - connects HTTP proxy server with gRPC server

All in all it's interlayer which redirects REST API requests to gRPC service to processing it