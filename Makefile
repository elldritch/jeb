KRPC_SERVICES={Drawing,InfernalRobotics,KerbalAlarmClock,RemoteTech,SpaceCenter,UI}
KSP_PATH=~/.steam/steam/steamapps/common/Kerbal\ Space\ Program/

.PHONY: all
all: services proto
	go install ./...

.PHONY: services
services: vendor/krpc
	cd vendor/krpc; bazel build //service/$(KRPC_SERVICES):ServiceDefinitions
	cp -f vendor/krpc/bazel-bin/service/$(KRPC_SERVICES)/*.json krpc/codegen

.PHONY: proto
proto: vendor/krpc
	cp vendor/krpc/protobuf/krpc.proto krpc/pb
	patch -p0 < krpc/pb/krpc.proto.patch
	protoc --go_out=$$GOPATH/src krpc/pb/krpc.proto

vendor/krpc:
	git submodule update --recursive --init
	ln -s $(KSP_PATH) ./vendor/krpc/lib/ksp

.PHONY: clean
clean:
	rm -rf vendor/krpc
	rm -f krpc/pb/*.proto
	rm -f krpc/codegen/*.json
	rm -f krpc/pb/krpc.pb.go
