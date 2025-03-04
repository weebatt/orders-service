protoc __with__ grpc-gateway:
```bash
protoc --go_out=./pkg/api/test --go_opt=paths=source_relative \
--go-grpc_out=./pkg/api/test  --go-grpc_opt=paths=source_relative \
--grpc-gateway_out=./pkg/api/test --grpc-gateway_opt=paths=source_relative \
./proto/api/*proto
```

Adding simple __docs__ of grpc endpoints with swagger:
```bash
 protoc -I . --openapiv2_out ./pkg/api/docs \                                            
./proto/api/*proto

```