package hclext

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/hashicorp/go-cty-funcs/crypto"
	"github.com/hashicorp/go-cty-funcs/filesystem"
	"github.com/hashicorp/go-cty-funcs/uuid"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"github.com/zclconf/go-cty/cty/gocty"
)

func NewFunctionMap() map[string]function.Function {

	return map[string]function.Function{

		"jsonencode":             stdlib.JSONEncodeFunc,
		"jsondecode":             stdlib.JSONDecodeFunc,
		"csvdecode":              stdlib.CSVDecodeFunc,
		"equal":                  stdlib.EqualFunc,
		"notequal":               stdlib.NotEqualFunc,
		"format":                 stdlib.FormatFunc,
		"join":                   stdlib.JoinFunc,
		"merge":                  stdlib.MergeFunc,
		"length":                 stdlib.LengthFunc,
		"keys":                   stdlib.KeysFunc,
		"values":                 stdlib.ValuesFunc,
		"flatten":                stdlib.FlattenFunc,
		"contains":               stdlib.ContainsFunc,
		"index":                  stdlib.IndexFunc,
		"lookup":                 stdlib.LookupFunc,
		"element":                stdlib.ElementFunc,
		"slice":                  stdlib.SliceFunc,
		"compact":                stdlib.CompactFunc,
		"distinct":               stdlib.DistinctFunc,
		"reverselist":            stdlib.ReverseListFunc,
		"setproduct":             stdlib.SetProductFunc,
		"setunion":               stdlib.SetUnionFunc,
		"setintersection":        stdlib.SetIntersectionFunc,
		"sethaselement":          stdlib.SetHasElementFunc,
		"setsubtract":            stdlib.SetSubtractFunc,
		"setsymmetricdifference": stdlib.SetSymmetricDifferenceFunc,
		"formatdate":             stdlib.FormatDateFunc,
		"timeadd":                stdlib.TimeAddFunc,
		"add":                    stdlib.AddFunc,
		"assertnotnull":          stdlib.AssertNotNullFunc,
		"byteslen":               stdlib.BytesLenFunc,
		"byteslice":              stdlib.BytesSliceFunc,
		"not":                    stdlib.NotFunc,
		"and":                    stdlib.AndFunc,
		"or":                     stdlib.OrFunc,
		"upper":                  stdlib.UpperFunc,
		"lower":                  stdlib.LowerFunc,
		"replace":                stdlib.ReplaceFunc,
		"split":                  stdlib.SplitFunc,
		"substr":                 stdlib.SubstrFunc,
		"trimprefix":             stdlib.TrimPrefixFunc,
		"trimsuffix":             stdlib.TrimSuffixFunc,
		"trimspace":              stdlib.TrimSpaceFunc,
		"trim":                   stdlib.TrimFunc,
		"chomp":                  stdlib.ChompFunc,
		"chunklist":              stdlib.ChunklistFunc,
		"coalesce":               stdlib.CoalesceFunc,
		"indent":                 stdlib.IndentFunc,
		"title":                  stdlib.TitleFunc,
		"abs":                    stdlib.AbsoluteFunc,
		"ceil":                   stdlib.CeilFunc,
		"div":                    stdlib.DivideFunc,
		"mod":                    stdlib.ModuloFunc,
		"floor":                  stdlib.FloorFunc,
		"max":                    stdlib.MaxFunc,
		"min":                    stdlib.MinFunc,
		"mul":                    stdlib.MultiplyFunc,
		"gte":                    stdlib.GreaterThanOrEqualToFunc,
		"gt":                     stdlib.GreaterThanFunc,
		"lte":                    stdlib.LessThanOrEqualToFunc,
		"lt":                     stdlib.LessThanFunc,
		"sub":                    stdlib.SubtractFunc,
		"neg":                    stdlib.NegateFunc,
		"int":                    stdlib.IntFunc,
		"log":                    stdlib.LogFunc,
		"pow":                    stdlib.PowFunc,
		"signum":                 stdlib.SignumFunc,
		"parseint":               stdlib.ParseIntFunc,
		"range":                  stdlib.RangeFunc,
		"formatlist":             stdlib.FormatListFunc,
		"regex":                  stdlib.RegexFunc,
		"regexall":               stdlib.RegexAllFunc,
		"regexreplace":           stdlib.RegexReplaceFunc,
		"zipmap":                 stdlib.ZipmapFunc,
		"coelscelist":            stdlib.CoalesceListFunc,
		"reverse":                stdlib.ReverseFunc,
		"sort":                   stdlib.SortFunc,

		"tomlencode":   tomlencodeFunc(),
		"tomldecode":   tomldecodeFunc(),
		"base64encode": base64encodeFunc(),
		"base64decode": base64decodeFunc(),
		"filepath":     filepathFunc(),

		"yamlencode": ctyyaml.YAMLEncodeFunc,
		"yamldecode": ctyyaml.YAMLDecodeFunc,
		"uuidv4":     uuid.V4Func,
		"uuidv5":     uuid.V5Func,
		"sha256":     crypto.Sha256Func,
		"sha512":     crypto.Sha512Func,
		"sha1":       crypto.Sha1Func,
		"md5":        crypto.Md5Func,
		"fileexists": filesystem.MakeFileExistsFunc(""),
		"fileread":   filesystem.MakeFileFunc("", false),
		"globfiles":  filesystem.MakeFileSetFunc(""),
	}
}

func base64encodeFunc() function.Function {
	return function.New(&function.Spec{
		Description: `Returns the Base64-encoded version of the given string.`,
		Params: []function.Parameter{
			{
				Name:             "str",
				Type:             cty.String,
				AllowUnknown:     false,
				AllowDynamicType: false,
				AllowNull:        false,
			},
		},
		Type: function.StaticReturnType(cty.String),
		// RefineResult: refineNonNull,
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			if len(args) != 1 {
				return cty.NilVal, fmt.Errorf("expected 1 argument, got %d", len(args))
			}
			if args[0].IsNull() {
				return cty.StringVal(""), nil
			}

			if args[0].Type() != cty.String {
				return cty.NilVal, fmt.Errorf("expected string, got %s", args[0].GoString())
			}
			return cty.StringVal(base64.StdEncoding.EncodeToString([]byte(args[0].AsString()))), nil
		},
	})
}

func base64decodeFunc() function.Function {
	return function.New(&function.Spec{
		Description: `Returns the Base64-decoded version of the given string.`,
		Params: []function.Parameter{
			{
				Name:             "str",
				Type:             cty.String,
				AllowUnknown:     false,
				AllowDynamicType: false,
				AllowNull:        false,
			},
		},
		Type: function.StaticReturnType(cty.String),
		// RefineResult: refineNonNull,
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			if len(args) != 1 {
				return cty.NilVal, fmt.Errorf("expected 1 argument, got %d", len(args))
			}
			if args[0].IsNull() {
				return cty.StringVal(""), nil
			}
			if args[0].Type() != cty.String {
				return cty.NilVal, fmt.Errorf("expected string, got %s", args[0].GoString())
			}
			dec, err := base64.StdEncoding.DecodeString(args[0].AsString())
			if err != nil {
				return cty.NilVal, err
			}
			return cty.StringVal(string(dec)), nil
		},
	})
}

func tomlencodeFunc() function.Function {
	return function.New(&function.Spec{
		Description: `Returns the TOML-encoded version of the given value.`,
		Params: []function.Parameter{
			{
				Name:             "value",
				Type:             cty.DynamicPseudoType,
				AllowUnknown:     false,
				AllowDynamicType: true,
				AllowNull:        true,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			if len(args) != 1 {
				return cty.NilVal, fmt.Errorf("expected 1 argument, got %d", len(args))
			}

			val := args[0]
			if val.IsNull() {
				return cty.StringVal(""), nil
			}

			var goVal interface{}
			err = gocty.FromCtyValue(val, &goVal)
			if err != nil {
				return cty.NilVal, err
			}

			var buf bytes.Buffer
			err = toml.NewEncoder(&buf).Encode(goVal)
			if err != nil {
				return cty.NilVal, err
			}

			return cty.StringVal(buf.String()), nil
		},
	})
}

func tomldecodeFunc() function.Function {
	return function.New(&function.Spec{
		Description: `Parses the given TOML string and returns a representation of its data.`,
		Params: []function.Parameter{
			{
				Name:             "str",
				Type:             cty.String,
				AllowUnknown:     false,
				AllowDynamicType: false,
				AllowNull:        false,
			},
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			if len(args) != 1 {
				return cty.NilVal, fmt.Errorf("expected 1 argument, got %d", len(args))
			}

			if args[0].IsNull() {
				return cty.NullVal(cty.DynamicPseudoType), nil
			}

			if args[0].Type() != cty.String {
				return cty.NilVal, fmt.Errorf("expected string, got %s", args[0].GoString())
			}

			var result interface{}
			err = toml.Unmarshal([]byte(args[0].AsString()), &result)
			if err != nil {
				return cty.NilVal, err
			}

			ctyVal, err := gocty.ToCtyValue(result, cty.DynamicPseudoType)
			if err != nil {
				return cty.NilVal, err
			}

			return ctyVal, nil
		},
	})
}

func filepathFunc() function.Function {
	return function.New(&function.Spec{
		Description: `Returns the joined path of the given path and the given path, using os specific path separator - accepts any number of arguments`,
		Params: []function.Parameter{
			{
				Name:             "path",
				Type:             cty.String,
				AllowUnknown:     true,
				AllowDynamicType: true,
				AllowNull:        true,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			if len(args) < 1 {
				return cty.NilVal, fmt.Errorf("expected at least 1 argument, got %d", len(args))
			}
			paths := make([]string, len(args))
			for i, arg := range args {
				paths[i] = arg.AsString()
			}
			return cty.StringVal(filepath.Join(paths...)), nil
		},
	})
}
