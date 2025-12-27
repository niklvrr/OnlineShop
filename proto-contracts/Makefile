.PHONY: order payment proto
order:
	protoc -I order order/order-service.proto --go_out=./gen/go/order --go_opt=paths=source_relative --go-grpc_out=./gen/go/order --go-grpc_opt=paths=source_relative
payment:
	protoc -I payment payment/payment-service.proto --go_out=./gen/go/payment --go_opt=paths=source_relative --go-grpc_out=./gen/go/payment --go-grpc_opt=paths=source_relative

proto: order payment
