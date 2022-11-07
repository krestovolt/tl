package main

import (
	"fmt"
	"os"
	"text/template"

	"github.com/gotd/tl"
)

func main() {
	fin, err := os.Open("./.out/schema.core_types.tl")
	if fin != nil {
		defer fin.Close()
	}
	if err != nil {
		panic(err)
	}

	schema, err := tl.Parse(fin)
	if err != nil {
		panic(err)
	}

	fout, err := os.Create("./.out/schema.core_types.out")
	if fout != nil {
		defer fout.Close()
	}
	if err != nil {
		panic(err)
	}

	// to tracks which definition using/depends on this type
	typeUsage := map[string][]string{}
	// e.g. Bool type with TL_boolFalse and TL_boolTrue
	for _, d := range schema.Definitions {
		typeName := d.Definition.Type.Name
		defs, exists := typeUsage[typeName]
		if !exists {
			defs = []string{d.Definition.Name}
		} else {
			defs = append(defs, d.Definition.Name)
		}
		typeUsage[typeName] = defs

		fmt.Fprintf(fout, "// %s: %s, Definition: %s, int32(crc32): %d\n",
			d.Category,
			d.Definition.Type.Name,
			d.Definition.Name,
			int32(d.Definition.ID),
		)
		fmt.Fprintln(fout, d.Definition.String())
		fmt.Fprintln(fout, "")
	}
	///////////////////

	funcMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"add": func(a int, b int) int {
			return a + b
		},
		"print_type": func(pos, len int, param *tl.Parameter) string {
			typename := param.Type.Name
			switch typename {
			case "int":
				typename = "int32"
			case "long":
				typename = "int64"
			}

			sep := "\n"
			if pos == len-1 {
				sep = ""
			}
			return fmt.Sprintf("%s %s = %d;%s", typename, param.Name, pos+3, sep)
		},
	}

	tpout, err := os.Create("./.out/schema.tl.core_types.proto")
	if tpout != nil {
		defer tpout.Close()
	}
	if err != nil {
		panic(err)
	}

	t, err := template.New("").Funcs(funcMap).ParseFiles("./generator/template-proto/schema.tl.core_types.tpl")
	if err != nil {
		panic(err)
	}

	loadedType := map[string]struct{}{}
	for i, def := range schema.Definitions {
		typeName := def.Definition.Type.Name
		_, typeLoaded := loadedType[typeName]
		if !typeLoaded {
			loadedType[def.Definition.Type.Name] = struct{}{}
		}

		err = t.ExecuteTemplate(tpout, "schema.tl.core_types.tpl", struct {
			WithPreamble bool
			LoadType     bool
			UsedBy       []string
			Definition   *tl.Definition
		}{
			WithPreamble: i == 0,
			LoadType:     !typeLoaded,
			UsedBy:       typeUsage[typeName],
			Definition:   &def.Definition,
		})
		if err != nil {
			panic(err)
		}
	}

	// x, _ := hex.DecodeString("a2813660")
	// i := binary.BigEndian.Uint32(x)
	// x := int(1431132616)
	// fmt.Fprintf(fout, "%x\n", uint32(x))
}
