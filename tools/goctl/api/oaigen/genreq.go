package oaigen

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
)

func genReqBody(route spec.Route) map[string]interface{} {
	resp := make(map[string]interface{})

	reqParams := route.RequestType.(spec.DefineStruct)
	bodyParams := reqParams.GetFillBodyMembers()
	formParams := reqParams.GetFormMembers()
	// embed the request body
	if len(bodyParams) == 0 && len(formParams) == 0 {
		return resp
	}
	if len(bodyParams) > 0 {
		bodyValue := DisplayType(reqParams)
		if len(formParams) > 0 {
			bodyValue = map[string]interface{}{
				"type": "object",
			}

			parameters := make(map[string]interface{})
			for _, param := range bodyParams {
				paramSchema := map[string]interface{}{
					"type": DisplayType(param.Type),
				}
				parameters[strings.ToLower(param.Name)] = paramSchema
			}
			bodyValue["properties"] = parameters
		}

		resp["requestBody"] = map[string]interface{}{
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": bodyValue,
				},
			},
		}
	}
	if len(formParams) > 0 {
		parameters := make([]map[string]interface{}, 0, len(formParams))
		for _, param := range formParams {
			paramMap := map[string]interface{}{
				"name":     strings.ToLower(param.Name),
				"in":       "query",
				"required": !param.IsOptional(),
				"schema":   DisplayType(param.Type),
			}
			parameters = append(parameters, paramMap)
		}
		resp["parameters"] = parameters
	}

	return resp
}

func genRespBody(route spec.Route) map[string]interface{} {
	resp := make(map[string]interface{})
	resp["responses"] = map[string]interface{}{
		"200": map[string]interface{}{
			"description": "Successful operation",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": DisplayType(route.ResponseType),
				},
			},
		},
	}
	return resp
}

// DisplayType display type
func DisplayType(type_ spec.Type) map[string]interface{} {
	switch v := type_.(type) {
	case spec.ArrayType:
		return map[string]interface{}{
			"type":  "array",
			"items": DisplayType(v.Value),
		}
	case spec.MapType:
		return map[string]interface{}{
			"type":                 "object",
			"additionalProperties": DisplayType(v.Value),
		}
	case spec.InterfaceType:
		return map[string]interface{}{
			"type": "object",
		}
	case spec.PointerType:
		return DisplayType(v.Type)
	case spec.PrimitiveType:
		swaggerType := "string"
		switch v.RawName {
		case "int", "int32", "int64":
			swaggerType = "integer"
			return map[string]interface{}{"type": swaggerType, "format": v.RawName}
		case "bool":
			swaggerType = "boolean"
		}
		return map[string]interface{}{"type": swaggerType}
	case spec.DefineStruct:
		return map[string]interface{}{
			"$ref": fmt.Sprintf("#/components/schemas/%s", v.Name()),
		}
	default:
		return map[string]interface{}{
			"type":   "unknown",
			"format": type_.Name(),
		}
	}
}
