package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"reflect"
	"regexp"
	"strings"
)

func main() {

	dec := json.NewDecoder(os.Stdin)

	var json interface{}
	err := dec.Decode(&json)
	if err != nil {
                panic(err)
	}

	fieldList := createFieldListFromMap(reflect.ValueOf(json))

	structType := &ast.StructType{
		Fields: &ast.FieldList{
			List: fieldList,
		},
	}

	typeSpec := &ast.TypeSpec{
		Name: ast.NewIdent("XXX"),
		Type: structType,
	}

	genDecl := &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{typeSpec},
	}

	cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	fset := token.NewFileSet()
	var buf bytes.Buffer
	err = cfg.Fprint(&buf, fset, genDecl)
	if err != nil {
		panic(err)
	}
	fmt.Print(buf.String())
}

func createFieldListFromMap(data reflect.Value) []*ast.Field {
	if data.Kind() != reflect.Map {
		panic("shoud be map type")
	}
	var fieldList []*ast.Field

	for _, key := range data.MapKeys() {
		value := data.MapIndex(key)
		var field *ast.Field

		if value.IsNil() {
			field = createField(
				ast.NewIdent("<nil>"),
				key.String(),
			)
		} else {
			switch value.Elem().Type().Kind() {
			case reflect.Map:
				_fieldList := createFieldListFromMap(value.Elem())

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
					typeName = "interface"
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
				panic(fmt.Sprintf("non support type", key.String(), "=>", value.Elem().Type().Kind().String()))
			}
		}
		fieldList = append(fieldList, field)
	}

	return fieldList
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
