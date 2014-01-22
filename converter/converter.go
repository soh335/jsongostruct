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
		var field *ast.Field

		if value.IsNil() {
			field = createField(
				ast.NewIdent("interface{}"),
				key.String(),
			)
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

				field = createField(
					structType,
					key.String(),
				)
			case reflect.Array, reflect.Slice:
				//TODO: detect struct ?
				// current detect slice type from head element
				var typeName string
				if value.Elem().Len() > 0 {
					typeName = value.Elem().Index(0).Elem().Type().String()
				} else {
					typeName = "interface{}"
				}
				field = createField(
					ast.NewIdent("[]"+typeName),
					key.String(),
				)
			case reflect.String, reflect.Bool, reflect.Float64:
				field = createField(
					ast.NewIdent(value.Elem().Type().Kind().String()),
					key.String(),
				)
			default:
				return nil, errors.New(fmt.Sprintf("non support type", key.String(), "=>", value.Elem().Type().Kind().String()))
			}
		}
		fieldList = append(fieldList, field)
	}

	return fieldList, nil
}

func createField(expr ast.Expr, name string) *ast.Field {

	//TODO: handling tag for struct
	return &ast.Field{
		Type:  expr,
		Names: []*ast.Ident{ast.NewIdent(camelize(name))},
		Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"" + name + "\"`"},
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
