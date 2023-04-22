//go:generate protoc -I ../../df-snapshot-protospec/api/protobuf-spec --go_out=plugins=grpc:../../ snapshot.proto

package snapshots

import (
	"context"
	"encoding/json"
	//"errors"
	"fmt"
	"net/url"
	"strings"

	proto "gate/g/rpc/snapshot"
	//"github.com/go-kit/kit/tracing/opentracing"
	kitoc "github.com/go-kit/kit/tracing/opencensus"

	"github.com/go-kit/kit/transport/http/jsonrpc"
)

func NewJSONRPCClient(instance string) (*Set, error) {

	// Quickly sanitize the instance string.
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	fmt.Printf("init snapshot client: %s\n", u)
	set := &Set{}

	set.SaveSnapshotEndpoint = jsonrpc.NewClient(
		u,
		"SaveSnapshot",
		jsonrpc.ClientRequestEncoder(encodeSnapshotRequest),
		jsonrpc.ClientResponseDecoder(decodeSaveSnapshotResponse),
		kitoc.JSONRPCClientTrace(),
	).Endpoint()

	return set, nil

}

func encodeSnapshotRequest(_ context.Context, obj interface{}) (json.RawMessage, error) {
	req, ok := obj.(*proto.SnapshotRequest)
	if !ok {
		return nil, fmt.Errorf("couldn't assert request as SnapshotRequest, got %T", obj)
	}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal request: %s", err)
	}
	return b, nil
}

func decodeSaveSnapshotResponse(_ context.Context, res jsonrpc.Response) (interface{}, error) {

	if res.Error != nil {
		return nil, *res.Error
	}
	var r *proto.SaveSnapshotResponse
	err := json.Unmarshal(res.Result, &r)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal body to SaveSnapshotResponse: %s", err)
	}
	return r, nil
}
