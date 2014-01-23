package converter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"reflect"
	"regexp"
	"strings"
)

func JsonGoStruct(r io.Reader, w io.Writer, name string) error {
	dec := json.NewDecoder(r)

	var json interface{}
	err := dec.Decode(&json)
	if err != nil {
		return err
	}

	fieldList, err := createFieldListFromMap(reflect.ValueOf(json))
	if err != nil {
		return err
	}

	structType := &ast.StructType{
		Fields: &ast.FieldList{
			List: fieldList,
		},
	}

	typeSpec := &ast.TypeSpec{
		Name: ast.NewIdent(name),
		Type: structType,
	}

	genDecl := &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{typeSpec},
	}

	cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	fset := token.NewFileSet()

	return cfg.Fprint(w, fset, genDecl)
}

func createFieldListFromMap(data reflect.Value) ([]*ast.Field, error) {
	if data.Kind() != reflect.Map {
		return nil, errors.New("shoud be map type")
	}
	var fieldList []*ast.Field

	for _, key := range data.MapKeys() {
		value := data.MapIndex(key)
		typ, err := createType(value)
		if err != nil {
			return nil, err
		}
		field := createField(typ, key.String())
		fieldList = append(fieldList, field)
	}

	return fieldList, nil
}

func createField(expr ast.Expr, name string) *ast.Field {
	return &ast.Field{
		Type:  expr,
		Names: []*ast.Ident{ast.NewIdent(camelize(name))},
		Tag:   &ast.BasicLit{ValuePos: 8, Kind: token.STRING, Value: "`json:\"" + name + "\"`"},
	}
}

func createType(value reflect.Value) (ast.Expr, error) {
	if value.IsNil() {
		// empty interface{}
		return &ast.InterfaceType{Methods: &ast.FieldList{Opening: 1, Closing: 2}}, nil
	} else {
		switch value.Elem().Type().Kind() {
		case reflect.Map:
			_fieldList, err := createFieldListFromMap(value.Elem())
			if err != nil {
				return nil, err
			}

			structType := &ast.StructType{
				Fields: &ast.FieldList{
					List: _fieldList,
				},
			}

			return structType, nil
		case reflect.Array, reflect.Slice:
			var elt ast.Expr
			var err error
			if value.Elem().Len() > 0 {
				elt, err = createType(value.Elem().Index(0))
				if err != nil {
					return nil, err
				}
			} else {
				elt = &ast.InterfaceType{Methods: &ast.FieldList{Opening: 1, Closing: 2}}
			}
			return &ast.ArrayType{Elt: elt}, nil
		case reflect.String, reflect.Bool, reflect.Float64:
			return ast.NewIdent(value.Elem().Type().Kind().String()), nil
		default:
			return nil, errors.New(fmt.Sprint("non support type", value.Elem().Type().Kind().String()))
		}
	}
}

func camelize(str string) string {
	if len(str) < 2 {
		return strings.ToTitle(str)
	}

	var list [][]byte
	bts := []byte(str)

	// head
	rehead := regexp.MustCompile(`^[0-9A-Za-z]+`)
	if m := rehead.Find(bts); m != nil {
		// head totitle, rest tolower.
		list = append(list, bytes.ToTitle([]byte{m[0]}), bytes.ToLower(m[1:]))
	}

	// rest
	re := regexp.MustCompile(`_([0-9A-Za-z]+)`)
	for _, m := range re.FindAllSubmatch(bts, -1) {
		list = append(list, bytes.ToTitle([]byte{m[1][0]}), bytes.ToLower(m[1][1:]))
	}

	return string(bytes.Join(list, nil))
}
