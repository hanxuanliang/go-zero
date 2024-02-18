package oaigen

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/tools/goctl/api/parser"
	apiutil "github.com/zeromicro/go-zero/tools/goctl/api/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
)

var (
	// VarStringDir describes the directory.
	VarStringDir string
	// VarStringAPI describes the API.
	VarStringAPI string
)

//go:embed oai.tpl
var markdownTemplate string

// OaiCommand gen openapi json file from command line
func OaiCommand(_ *cobra.Command, _ []string) error {
	apiFile := VarStringAPI
	dir := VarStringDir

	if len(apiFile) == 0 {
		return errors.New("missing -api")
	}
	if len(dir) == 0 {
		return errors.New("missing -dir")
	}

	return DoGenOpenapiJson(apiFile, dir)
}

func DoGenOpenapiJson(apiFile, dir string) error {
	api, err := parser.Parse(apiFile)
	if err != nil {
		return err
	}

	if err := api.Validate(); err != nil {
		return err
	}

	logx.Must(pathx.MkdirIfNotExist(dir))

	fp, _, err := apiutil.MaybeCreateFile(dir, "", "openapi.json")
	if err != nil {
		return err
	}
	defer fp.Close()

	serviceMap := genService(api)
	var (
		builder   strings.Builder
		tmplBytes bytes.Buffer
	)
	infoStr, _ := json.MarshalIndent(serviceMap["info"], "", "    ")
	pathStr, _ := json.MarshalIndent(serviceMap["paths"], "", "    ")
	componentsStr, _ := json.MarshalIndent(serviceMap["components"], "", "    ")

	oaiT := template.Must(template.New("markdownTemplate").Parse(markdownTemplate))
	if err := oaiT.Execute(&tmplBytes, map[string]string{
		"info":       string(infoStr),
		"paths":      string(pathStr),
		"components": string(componentsStr),
	}); err != nil {
		return err
	}

	builder.Write(tmplBytes.Bytes())
	_, err = fp.WriteString(strings.Replace(builder.String(), "&#34;", `"`, -1))
	return err
}

func genService(api *spec.ApiSpec) map[string]interface{} {
	res := map[string]interface{}{
		"openapi": "3.0.0",
		"paths":   genRoute(api.Service),
		"info": map[string]interface{}{
			"title":       api.Info.Properties["title"],
			"description": api.Info.Properties["desc"],
			"version":     api.Info.Properties["version"],
			"contact": map[string]interface{}{
				"name":  api.Info.Properties["author"],
				"email": api.Info.Properties["email"],
			},
		},
	}

	res = mergeMap(res, genSchema(api.Types))
	return res
}

func genRoute(service spec.Service) map[string]interface{} {
	res := make(map[string]interface{})

	for _, group := range service.Groups {
		for _, route := range group.Routes {
			path := route.Path
			prefix := group.Annotation.Properties[spec.RoutePrefixKey]
			if prefix != "" {
				path = fmt.Sprintf("/%s/%s", prefix, route.Path)
			}

			reqBody := genReqBody(route)
			respBody := genRespBody(route)
			operation := map[string]interface{}{
				"tags":        []string{group.Annotation.Properties["group"]},
				"parameters":  reqBody["parameters"],
				"requestBody": reqBody["requestBody"],
				"responses":   respBody["responses"],
			}

			res[path] = map[string]interface{}{
				strings.ToLower(route.Method): operation,
			}
		}
	}

	return res
}

func genReqBody(route spec.Route) map[string]interface{} {
	resp := make(map[string]interface{})

	reqParams := route.RequestType.(spec.DefineStruct)
	bodyParams := reqParams.GetFillBodyMembers()
	formParams := reqParams.GetFormMembers()
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
				parameters[strings.ToLower(param.Name)] = DisplayType(param.Type)
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

func genSchema(types []spec.Type) map[string]interface{} {
	schema := make(map[string]interface{})
	for _, tp := range types {
		structType, ok := tp.(spec.DefineStruct)
		if !ok {
			fmt.Println(fmt.Errorf("unspport struct type: %s", tp.Name()))
			continue
		}

		memberSchema := make(map[string]interface{})
		for _, m := range structType.Members {
			memberSchema[strings.ToLower(m.Name)] = DisplayType(m.Type)
		}
		schema[structType.RawName] = map[string]interface{}{
			"type":       "object",
			"properties": memberSchema,
		}
	}
	return map[string]interface{}{
		"components": map[string]interface{}{
			"schemas": schema,
		},
	}
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

func mergeMap(dst, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
