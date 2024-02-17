package oaigen

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	apiParser "github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/parser"
	"reflect"
	"testing"
)

func TestDisplayType(t *testing.T) {
	type args struct {
		type_ spec.Type
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "test array",
			args: args{
				type_: &spec.ArrayType{
					RawName: "[]int64",
					Value:   &spec.PrimitiveType{RawName: "int64"},
				},
			},
			want: map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "integer", "format": "int64"},
			},
		},
		{
			name: "test struct",
			args: args{
				type_: &spec.DefineStruct{
					RawName: "User",
					Members: []spec.Member{
						{
							Name: "id",
							Type: &spec.PrimitiveType{RawName: "int"},
						},
						{
							Name: "name",
							Type: &spec.PrimitiveType{RawName: "string"},
						},
					},
				},
			},
			want: map[string]interface{}{
				"$ref": "#/components/schemas/User",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DisplayType(tt.args.type_); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DisplayType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_genRespBody(t *testing.T) {
	type args struct {
		source string
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "test genRespBody",
			args: args{
				source: "./testdata/example-dev.api",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiSpec, err := apiParser.Parse(tt.args.source, "")
			assert.Nil(t, err)

			respBody := genRespBody(apiSpec.Service.Routes()[0])
			reqBody := genReqBody(apiSpec.Service.Routes()[0])

			reqStr, _ := json.Marshal(reqBody)
			println("reqStr", string(reqStr))
			respStr, _ := json.Marshal(respBody)
			println("respStr", string(respStr))
		})
	}
}
