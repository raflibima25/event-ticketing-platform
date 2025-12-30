.PHONY: proto-payment proto-ticketing proto-notification proto-all clean

proto-payment:
	mkdir -p pb/payment
	protoc --proto_path=proto \
		--go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		proto/payment/payment.proto

proto-ticketing:
	mkdir -p pb/ticketing
	protoc --proto_path=proto \
		--go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		proto/ticketing/ticketing.proto

proto-notification:
	mkdir -p pb/notification
	protoc --proto_path=proto \
		--go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		proto/notification/notification.proto

proto-all: proto-payment proto-ticketing proto-notification

clean:
	rm -rf pb
