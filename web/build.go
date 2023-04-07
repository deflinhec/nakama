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

package web

//go:generate protoc -I. -I../api -I../vendor -I../build/grpc-gateway-v2.3.0/third_party/googleapis -I../vendor/github.com/grpc-ecosystem/grpc-gateway/v2 --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=logtostderr=true --openapiv2_out=. --openapiv2_opt=logtostderr=true,allow_delete_body=true web.proto
//go:generate sh -c "(cd ../console/openapi-gen-angular && go run . -i '../../web/web.swagger.json' -o '../../web/ui/src/app/app.service.ts' -rm_prefix='application,nakamaapplication,nakama,web,Application_')"
