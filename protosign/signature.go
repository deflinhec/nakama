// Copyright 2023 Deflinhec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protosign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// JsonFmt returns a signature base on sorted json key fields.
func JsonFmt(msg interface{}, key string) ([]byte, error) {
	jsonsort := func(b []byte) ([]byte, error) {
		var ifce interface{}
		err := json.Unmarshal(b, &ifce)
		if err != nil {
			return nil, err
		}
		return json.Marshal(ifce)
	}
	var hash []byte
	h := hmac.New(sha256.New, []byte(key))
	if m, ok := msg.(protoreflect.ProtoMessage); !ok {
		return nil, errors.New("casting request to protoreflect.Message")
	} else if b, err := (protojson.MarshalOptions{
		UseProtoNames: false, UseEnumNumbers: true, EmitUnpopulated: true,
	}).Marshal(m); err != nil {
		return nil, errors.New("marshalling message")
	} else if b, err = jsonsort(b); err != nil {
		return nil, errors.New("marshalling message")
	} else if _, err = h.Write(b); err != nil {
		return nil, errors.New("writing response bytes to hasher")
	} else {
		hash = h.Sum(nil)
	}
	return hash, nil
}

// ProtoFmt returns a signature base on protobuf binary.
func ProtoFmt(msg interface{}, key string) ([]byte, error) {
	var hash []byte
	h := hmac.New(sha256.New, []byte(key))
	if m, ok := msg.(protoreflect.ProtoMessage); !ok {
		return nil, errors.New("casting request to protoreflect.Message")
	} else if b, err := proto.Marshal(m); err != nil {
		return nil, errors.New("marshalling message")
	} else if _, err = h.Write(b); err != nil {
		return nil, errors.New("marshalling message")
	} else {
		hash = h.Sum(nil)
	}
	return hash, nil
}

func Gen(msg interface{}, key string) ([]string, error) {
	signatures := make([]string, 2)
	if b, err := JsonFmt(msg, key); err != nil {
		return nil, err
	} else {
		signatures[0] = hex.EncodeToString(b)
	}
	if b, err := ProtoFmt(msg, key); err != nil {
		return nil, err
	} else {
		signatures[1] = hex.EncodeToString(b)
	}
	return signatures, nil
}

func Sign(md *metadata.MD, msg interface{}, key string) error {
	signatures, err := Gen(msg, key)
	if err != nil {
		return err
	}
	for _, signature := range signatures {
		md.Append("X-Signature", signature)
	}
	return nil
}

func Verify(md *metadata.MD, msg interface{}, key string) (bool, error) {
	b, err := JsonFmt(msg, key)
	if err != nil {
		return false, err
	}
	signature := hex.EncodeToString(b)
	signatures := md.Get("X-Signature")
	for _, s := range signatures {
		if s == signature {
			return true, nil
		}
	}
	return false, nil
}
