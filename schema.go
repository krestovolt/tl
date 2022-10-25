package tl

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// SchemaDefinition is annotated Definition with Category.
type SchemaDefinition struct {
	Scope       ScopeTypeEnum `json:"scope,omitempty"`       // scope (comments)
	Annotations []Annotation  `json:"annotations,omitempty"` // annotations (comments)
	Definition  Definition    `json:"definition"`            // definition
	Category    Category      `json:"category"`              // category of definition (function or type)
}

// Class describes a non-bare Type with one or more constructors.
//
// Example: `//@class InputChatPhoto @description Describes input chat photo`.
type Class struct {
	Name        string
	Description string
}

// Schema represents single TL file with information about definitions and
// so-called "Classes" aka non-bare types with one or multiple constructors.
type Schema struct {
	Layer       int                `json:"layer,omitempty"`
	Definitions []SchemaDefinition `json:"definitions"`
	Classes     []Class            `json:"classes,omitempty"`
}

// WriteTo writes whole schema to w, implementing io.WriterTo.
func (s Schema) WriteTo(w io.Writer) (int64, error) {
	classes := make(map[string]Class)
	classDefined := make(map[string]struct{})
	for _, class := range s.Classes {
		classes[class.Name] = class
	}

	category := CategoryType

	// Probably we can write to w directly, but schemas that are larger than
	// few megs are a problem itself.
	var b strings.Builder
	for _, d := range s.Definitions {
		if d.Category != category {
			category = d.Category
			b.WriteString("\n")
			switch category {
			case CategoryType:
				b.WriteString(tokTypes)
			case CategoryFunction:
				b.WriteString(tokFunctions)
			}
			b.WriteString("\n\n")
		}

		if class, exist := classes[d.Definition.Type.Name]; exist {
			// Describing class if not already defined.
			if _, defined := classDefined[class.Name]; !defined {
				b.WriteString(singleLineAnnotations([]Annotation{
					{Name: AnnotationClass, Value: class.Name},
					{Name: AnnotationDescription, Value: class.Description},
				}))
				classDefined[class.Name] = struct{}{}
				b.WriteString("\n\n")
			}
		}
		for _, a := range d.Annotations {
			b.WriteString(a.String())
			b.WriteString("\n")
		}
		// Writing definition itself.
		b.WriteString(d.Definition.String())
		b.WriteString(";\n\n")
	}

	if s.Layer != 0 {
		b.WriteString(fmt.Sprintf("// LAYER %d\n", s.Layer))
	}

	n, err := w.Write([]byte(b.String()))
	return int64(n), err
}

const (
	vectorDefinition       = "vector {t:Type} # [ t ] = Vector t;"
	vectorDefinitionWithID = "vector#1cb5c415 {t:Type} # [ t ] = Vector t;"
)

// Parse reads Schema from reader.
//
// Can return i/o or validation error.
func Parse(reader io.Reader) (*Schema, error) {
	var (
		def  SchemaDefinition
		line int

		category    = CategoryType
		schema      = &Schema{}
		scanner     = bufio.NewScanner(reader)
		activeScope = ScopeEmpty
	)
	for scanner.Scan() {
		line++
		s := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(s, "///") {
			s = s[1:] // normalize comments
		}
		switch s {
		case "":
			continue
		case tokFunctions:
			category = CategoryFunction
			continue
		case tokTypes:
			category = CategoryType
			continue
		case vectorDefinition, vectorDefinitionWithID:
			// Special case for vector.
			continue
		}
		if scope, ok := ParseScope(s); ok {
			// The activeScope will be always caried over the next loop
			// unless there is scope terminator "Scope End" (as defined by `tokScopeEnd`)
			activeScope = scope
			continue
		}
		if strings.HasPrefix(s, tokLayer) {
			// Layer version annotation.
			layer, err := strconv.Atoi(strings.TrimPrefix(s, tokLayer))
			if err != nil {
				return nil, fmt.Errorf("failed to parse layer: %w", err)
			}
			schema.Layer = layer
		}
		if strings.HasPrefix(s, "//@") {
			// Found annotation.
			ann, err := parseAnnotation(s)
			if err != nil {
				return nil, fmt.Errorf("failed to parse line %d: %w", line, err)
			}
			if strings.HasPrefix(s, "//@"+AnnotationClass) {
				// Handling class annotation as special case.
				var class Class
				for _, a := range ann {
					if a.Name == AnnotationClass {
						class.Name = a.Value
					}
					if a.Name == AnnotationDescription {
						class.Description = a.Value
					}
				}
				if class.Name != "" && class.Description != "" {
					schema.Classes = append(schema.Classes, class)
				}
				// Reset annotations so we don't include them to next type.
				ann = ann[:0]
			}

			def.Annotations = append(def.Annotations, ann...)
			continue // annotation is parsed, moving to next line
		}
		if strings.HasPrefix(s, "//") {
			continue // skip comments
		}

		// Type definition started.
		def.Category = category
		if err := def.Definition.Parse(s); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: definition: %w", line, err)
		}

		// Validating annotations.
		paramExist := map[string]struct{}{}
		for _, p := range def.Definition.Params {
			paramExist[p.Name] = struct{}{}
		}
		for _, ann := range def.Annotations {
			if ann.Name == AnnotationDescription {
				continue
			}
			searchFor := ann.Name
			if ann.Name == AnnotationParamDescription {
				// Special case for "description" parameter name that collides
				// with global description.
				searchFor = "description"
			}
			if _, ok := paramExist[searchFor]; !ok {
				// Probably such errors can be just skipped, but seems like it
				// is OK to consider this as hard failure.
				return nil, fmt.Errorf("failed to parse line %d: "+
					"can't find param for annotation %q", line, ann.Name)
			}
		}

		def.Scope = activeScope
		schema.Definitions = append(schema.Definitions, def)
		def = SchemaDefinition{} // reset definition
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan: %w", err)
	}

	// Remaining type.
	if def.Definition.ID != 0 {
		schema.Definitions = append(schema.Definitions, def)
	}

	return schema, nil
}
