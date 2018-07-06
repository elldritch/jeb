# Binary variables
GOBIN=$(GOPATH)/bin
ENSURE_DEPS=$(GOBIN)/ensure-deps
JEB=$(GOBIN)/jeb

# Target variables
KRPC_SERVICE_DEFINITIONS=krpc/codegen/*.json
KRPC_PROTOBUF=krpc/pb/krpc.pb.go

# Source variables
KRPC_SERVICES={Drawing,InfernalRobotics,KerbalAlarmClock,RemoteTech,SpaceCenter,UI}
KSP_PATH=~/.steam/steam/steamapps/common/Kerbal\ Space\ Program/
GO_SOURCES=$(shell find -type f -name \*.go | grep -v vendor)

# Build targets
$(JEB): $(KRPC_PROTOBUF) $(KRPC_SERVICE_DEFINITIONS) $(GO_SOURCES)
	go install ./...

$(KRPC_SERVICE_DEFINITIONS): vendor/krpc
	cd vendor/krpc; bazel build //service/$(KRPC_SERVICES):ServiceDefinitions; bazel shutdown
	cp -f vendor/krpc/bazel-bin/service/$(KRPC_SERVICES)/*.json krpc/codegen

$(KRPC_PROTOBUF): vendor/krpc
	cp vendor/krpc/protobuf/krpc.proto krpc/pb
	patch -p0 < krpc/pb/krpc.proto.patch
	protoc --go_out=$(GOPATH)/src krpc/pb/krpc.proto
	touch krpc/pb/krpc.pb.go

vendor/krpc:
	git submodule update --recursive --init
	ln -s $(KSP_PATH) ./vendor/krpc/lib/ksp

# Task targets
.PHONY:
check: $(ENSURE_DEPS)
	ensure-deps -exclude-import github.com/ilikebits/jeb

.PHONY: clean
clean:
	rm -rf vendor/krpc
	rm -f krpc/codegen/*.json
	rm -f krpc/pb/*.proto
	rm -f krpc/pb/*.pb.go

# Tool targets
$(ENSURE_DEPS):
	go get -u -v github.com/glerchundi/ensure-deps
