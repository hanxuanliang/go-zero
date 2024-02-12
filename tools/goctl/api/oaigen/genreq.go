package oaigen

import (
	_ "embed"
	"fmt"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
)

// DisplayType display type
func DisplayType(type_ spec.Type) map[string]interface{} {
	switch v := type_.(type) {
	case *spec.ArrayType:
		return map[string]interface{}{
			"type":  "array",
			"items": DisplayType(v.Value),
		}
	case *spec.MapType:
		return map[string]interface{}{
			"type":                 "object",
			"additionalProperties": DisplayType(v.Value),
		}
	case *spec.InterfaceType:
		return map[string]interface{}{
			"type": "object",
		}
	case *spec.PointerType:
		return DisplayType(v.Type)
	case *spec.PrimitiveType:
		swaggerType := "string"
		switch v.RawName {
		case "int", "int32", "int64":
			swaggerType = "integer"
			return map[string]interface{}{"type": swaggerType, "format": v.RawName}
		case "bool":
			swaggerType = "boolean"
		}
		return map[string]interface{}{"type": swaggerType}
	case *spec.DefineStruct:
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
