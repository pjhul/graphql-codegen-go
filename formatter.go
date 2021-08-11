package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

func NewFormatter(w io.Writer) *Formatter {
	return &Formatter{Writer: w}
}

type Formatter struct {
	Writer io.Writer

	indent      int
	emitBuiltin bool

	padNext  bool
	lineHead bool
}

func (f *Formatter) writeString(s string) {
	_, _ = f.Writer.Write([]byte(s))
}

func (f *Formatter) writeIndent() *Formatter {
	if f.lineHead {
		f.writeString(strings.Repeat("\t", f.indent))
	}
	f.lineHead = false
	f.padNext = false

	return f
}

func (f *Formatter) WriteNewline() *Formatter {
	f.writeString("\n")
	f.lineHead = true
	f.padNext = false

	return f
}

func (f *Formatter) WriteWord(word string) *Formatter {
	if f.lineHead {
		f.writeIndent()
	}
	if f.padNext {
		f.writeString(" ")
	}
	f.writeString(strings.TrimSpace(word))
	f.padNext = true

	return f
}

func (f *Formatter) WriteString(s string) *Formatter {
	if f.lineHead {
		f.writeIndent()
	}
	if f.padNext {
		f.writeString(" ")
	}
	f.writeString(s)
	f.padNext = false

	return f
}

func (f *Formatter) WriteDescription(s string) *Formatter {
	if s == "" {
		return f
	}

	f.WriteString(`"""`).WriteNewline()

	ss := strings.Split(s, "\n")
	for _, s := range ss {
		f.WriteString(s).WriteNewline()
	}

	f.WriteString(`"""`).WriteNewline()

	return f
}

func (f *Formatter) IncrementIndent() {
	f.indent++
}

func (f *Formatter) DecrementIndent() {
	f.indent--
}

func (f *Formatter) NoPadding() *Formatter {
	f.padNext = false

	return f
}

func (f *Formatter) NeedPadding() *Formatter {
	f.padNext = true

	return f
}

func (f *Formatter) FormatSchema(schema *ast.Schema) {
	if schema == nil {
		return
	}

	var inSchema bool
	startSchema := func() {
		if !inSchema {
			inSchema = true

			f.WriteWord("schema").WriteString("{").WriteNewline()
			f.IncrementIndent()
		}
	}
	if schema.Query != nil && schema.Query.Name != "Query" {
		startSchema()
		f.WriteWord("query").NoPadding().WriteString(":").NeedPadding()
		f.WriteWord(schema.Query.Name).WriteNewline()
	}
	if schema.Mutation != nil && schema.Mutation.Name != "Mutation" {
		startSchema()
		f.WriteWord("mutation").NoPadding().WriteString(":").NeedPadding()
		f.WriteWord(schema.Mutation.Name).WriteNewline()
	}
	if schema.Subscription != nil && schema.Subscription.Name != "Subscription" {
		startSchema()
		f.WriteWord("subscription").NoPadding().WriteString(":").NeedPadding()
		f.WriteWord(schema.Subscription.Name).WriteNewline()
	}
	if inSchema {
		f.DecrementIndent()
		f.WriteString("}").WriteNewline()
	}

	directiveNames := make([]string, 0, len(schema.Directives))
	for name := range schema.Directives {
		directiveNames = append(directiveNames, name)
	}
	sort.Strings(directiveNames)
	for _, name := range directiveNames {
		f.FormatDirectiveDefinition(schema.Directives[name])
	}

	typeNames := make([]string, 0, len(schema.Types))
	for name := range schema.Types {
		typeNames = append(typeNames, name)
	}
	sort.Strings(typeNames)
	for _, name := range typeNames {
		f.FormatDefinition(schema.Types[name], false)
	}
}

func (f *Formatter) FormatSchemaDocument(doc *ast.SchemaDocument) {
	// TODO emit by position based order

	if doc == nil {
		return
	}

	f.FormatSchemaDefinitionList(doc.Schema, false)
	f.FormatSchemaDefinitionList(doc.SchemaExtension, true)

	f.FormatDirectiveDefinitionList(doc.Directives)

	f.FormatDefinitionList(doc.Definitions, false)
	f.FormatDefinitionList(doc.Extensions, true)
}

func (f *Formatter) FormatQueryDocument(doc *ast.QueryDocument) {
	// TODO emit by position based order

	if doc == nil {
		return
	}

	f.FormatOperationList(doc.Operations)
	f.FormatFragmentDefinitionList(doc.Fragments)
}

func (f *Formatter) FormatSchemaDefinitionList(lists ast.SchemaDefinitionList, extension bool) {
	if len(lists) == 0 {
		return
	}

	if extension {
		f.WriteWord("extend")
	}
	f.WriteWord("schema").WriteString("{").WriteNewline()
	f.IncrementIndent()

	for _, def := range lists {
		f.FormatSchemaDefinition(def)
	}

	f.DecrementIndent()
	f.WriteString("}").WriteNewline()
}

func (f *Formatter) FormatSchemaDefinition(def *ast.SchemaDefinition) {
	f.WriteDescription(def.Description)

	f.FormatDirectiveList(def.Directives)

	f.FormatOperationTypeDefinitionList(def.OperationTypes)
}

func (f *Formatter) FormatOperationTypeDefinitionList(lists ast.OperationTypeDefinitionList) {
	for _, def := range lists {
		f.FormatOperationTypeDefinition(def)
	}
}

func (f *Formatter) FormatOperationTypeDefinition(def *ast.OperationTypeDefinition) {
	f.WriteWord(string(def.Operation)).NoPadding().WriteString(":").NeedPadding()
	f.WriteWord(def.Type)
	f.WriteNewline()
}

func (f *Formatter) FormatFieldList(fieldList ast.FieldList) {
	if len(fieldList) == 0 {
		return
	}

	f.WriteString("{").WriteNewline()
	f.IncrementIndent()

	for _, field := range fieldList {
		f.FormatFieldDefinition(field)
	}

	f.DecrementIndent()
	f.WriteString("}")
}

func (f *Formatter) FormatFieldDefinition(field *ast.FieldDefinition) {
	if !f.emitBuiltin && strings.HasPrefix(field.Name, "__") {
		return
	}

	f.WriteDescription(field.Description)

	f.WriteWord(field.Name).NoPadding()
	f.FormatArgumentDefinitionList(field.Arguments)
	f.NoPadding().WriteString(":").NeedPadding()
	f.FormatType(field.Type)

	if field.DefaultValue != nil {
		f.WriteWord("=")
		f.FormatValue(field.DefaultValue)
	}

	f.FormatDirectiveList(field.Directives)

	f.WriteNewline()
}

func (f *Formatter) FormatArgumentDefinitionList(lists ast.ArgumentDefinitionList) {
	if len(lists) == 0 {
		return
	}

	f.WriteString("(")
	for idx, arg := range lists {
		f.FormatArgumentDefinition(arg)

		if idx != len(lists)-1 {
			f.NoPadding().WriteWord(",")
		}
	}
	f.NoPadding().WriteString(")").NeedPadding()
}

func (f *Formatter) FormatArgumentDefinition(def *ast.ArgumentDefinition) {
	if def.Description != "" {
		f.WriteNewline().IncrementIndent()
		f.WriteDescription(def.Description)
	}

	f.WriteWord(def.Name).NoPadding().WriteString(":").NeedPadding()
	f.FormatType(def.Type)

	if def.DefaultValue != nil {
		f.WriteWord("=")
		f.FormatValue(def.DefaultValue)
	}

	if def.Description != "" {
		f.DecrementIndent()
		f.WriteNewline()
	}
}

func (f *Formatter) FormatDirectiveLocation(location ast.DirectiveLocation) {
	f.WriteWord(string(location))
}

func (f *Formatter) FormatDirectiveDefinitionList(lists ast.DirectiveDefinitionList) {
	if len(lists) == 0 {
		return
	}

	for _, dec := range lists {
		f.FormatDirectiveDefinition(dec)
	}
}

func (f *Formatter) FormatDirectiveDefinition(def *ast.DirectiveDefinition) {
	if !f.emitBuiltin {
		if def.Position.Src.BuiltIn {
			return
		}
	}

	f.WriteDescription(def.Description)
	f.WriteWord("directive").WriteString("@").WriteWord(def.Name)

	if len(def.Arguments) != 0 {
		f.NoPadding()
		f.FormatArgumentDefinitionList(def.Arguments)
	}

	if len(def.Locations) != 0 {
		f.WriteWord("on")

		for idx, dirLoc := range def.Locations {
			f.FormatDirectiveLocation(dirLoc)

			if idx != len(def.Locations)-1 {
				f.WriteWord("|")
			}
		}
	}

	f.WriteNewline()
}

func (f *Formatter) FormatDefinitionList(lists ast.DefinitionList, extend bool) {
	if len(lists) == 0 {
		return
	}

	for _, dec := range lists {
		f.FormatDefinition(dec, extend)
	}
}

func (f *Formatter) FormatDefinition(def *ast.Definition, extend bool) {
	if !f.emitBuiltin && def.BuiltIn {
		return
	}

	f.WriteDescription(def.Description)

	if extend {
		f.WriteWord("extend")
	}

	switch def.Kind {
	case ast.Scalar:
		f.WriteWord("scalar").WriteWord(def.Name)

	case ast.Object:
		f.WriteWord("type").WriteWord(def.Name)

	case ast.Interface:
		f.WriteWord("interface").WriteWord(def.Name)

	case ast.Union:
		f.WriteWord("union").WriteWord(def.Name)

	case ast.Enum:
		f.WriteWord("enum").WriteWord(def.Name)

	case ast.InputObject:
		f.WriteWord("input").WriteWord(def.Name)
	}

	if len(def.Interfaces) != 0 {
		f.WriteWord("implements").WriteWord(strings.Join(def.Interfaces, " & "))
	}

	f.FormatDirectiveList(def.Directives)

	if len(def.Types) != 0 {
		f.WriteWord("=").WriteWord(strings.Join(def.Types, " | "))
	}

	f.FormatFieldList(def.Fields)

	f.FormatEnumValueList(def.EnumValues)

	f.WriteNewline()
}

func (f *Formatter) FormatEnumValueList(lists ast.EnumValueList) {
	if len(lists) == 0 {
		return
	}

	f.WriteString("{").WriteNewline()
	f.IncrementIndent()

	for _, v := range lists {
		f.FormatEnumValueDefinition(v)
	}

	f.DecrementIndent()
	f.WriteString("}")
}

func (f *Formatter) FormatEnumValueDefinition(def *ast.EnumValueDefinition) {
	f.WriteDescription(def.Description)

	f.WriteWord(def.Name)
	f.FormatDirectiveList(def.Directives)

	f.WriteNewline()
}

func (f *Formatter) FormatOperationList(lists ast.OperationList) {
	for _, def := range lists {
		f.FormatOperationDefinition(def)
	}
}

func (f *Formatter) FormatOperationDefinition(def *ast.OperationDefinition) {
	f.WriteWord(string(def.Operation))
	if def.Name != "" {
		f.WriteWord(def.Name)
	}
	f.FormatVariableDefinitionList(def.VariableDefinitions)
	f.FormatDirectiveList(def.Directives)

	if len(def.SelectionSet) != 0 {
		f.FormatSelectionSet(def.SelectionSet)
		f.WriteNewline()
	}
}

func (f *Formatter) FormatDirectiveList(lists ast.DirectiveList) {
	if len(lists) == 0 {
		return
	}

	for _, dir := range lists {
		f.FormatDirective(dir)
	}
}

func (f *Formatter) FormatDirective(dir *ast.Directive) {
	f.WriteString("@").WriteWord(dir.Name)
	f.FormatArgumentList(dir.Arguments)
}

func (f *Formatter) FormatArgumentList(lists ast.ArgumentList) {
	if len(lists) == 0 {
		return
	}
	f.NoPadding().WriteString("(")
	for idx, arg := range lists {
		f.FormatArgument(arg)

		if idx != len(lists)-1 {
			f.NoPadding().WriteWord(",")
		}
	}
	f.WriteString(")").NeedPadding()
}

func (f *Formatter) FormatArgument(arg *ast.Argument) {
	f.WriteWord(arg.Name).NoPadding().WriteString(":").NeedPadding()
	f.WriteString(arg.Value.String())
}

func (f *Formatter) FormatFragmentDefinitionList(lists ast.FragmentDefinitionList) {
	for _, def := range lists {
		f.FormatFragmentDefinition(def)
	}
}

func (f *Formatter) FormatFragmentDefinition(def *ast.FragmentDefinition) {
	f.WriteWord("fragment").WriteWord(def.Name)
	f.FormatVariableDefinitionList(def.VariableDefinition)
	f.WriteWord("on").WriteWord(def.TypeCondition)
	f.FormatDirectiveList(def.Directives)

	if len(def.SelectionSet) != 0 {
		f.FormatSelectionSet(def.SelectionSet)
		f.WriteNewline()
	}
}

func (f *Formatter) FormatVariableDefinitionList(lists ast.VariableDefinitionList) {
	if len(lists) == 0 {
		return
	}

	f.WriteString("(")
	for idx, def := range lists {
		f.FormatVariableDefinition(def)

		if idx != len(lists)-1 {
			f.NoPadding().WriteWord(",")
		}
	}
	f.NoPadding().WriteString(")").NeedPadding()
}

func (f *Formatter) FormatVariableDefinition(def *ast.VariableDefinition) {
	f.WriteString("$").WriteWord(def.Variable).NoPadding().WriteString(":").NeedPadding()
	f.FormatType(def.Type)

	if def.DefaultValue != nil {
		f.WriteWord("=")
		f.FormatValue(def.DefaultValue)
	}

	// TODO https://github.com/vektah/gqlparser/v2/issues/102
	//   VariableDefinition : Variable : Type DefaultValue? Directives[Const]?
}

func (f *Formatter) FormatSelectionSet(sets ast.SelectionSet) {
	if len(sets) == 0 {
		return
	}

	f.WriteString("{").WriteNewline()
	f.IncrementIndent()

	for _, sel := range sets {
		f.FormatSelection(sel)
	}

	f.DecrementIndent()
	f.WriteString("}")
}

func (f *Formatter) FormatSelection(selection ast.Selection) {
	switch v := selection.(type) {
	case *ast.Field:
		f.FormatField(v)

	case *ast.FragmentSpread:
		f.FormatFragmentSpread(v)

	case *ast.InlineFragment:
		f.FormatInlineFragment(v)

	default:
		panic(fmt.Errorf("unknown Selection type: %T", selection))
	}

	f.WriteNewline()
}

func (f *Formatter) FormatField(field *ast.Field) {
	if field.Alias != "" && field.Alias != field.Name {
		f.WriteWord(field.Alias).NoPadding().WriteString(":").NeedPadding()
	}
	f.WriteWord(field.Name)

	if len(field.Arguments) != 0 {
		f.NoPadding()
		f.FormatArgumentList(field.Arguments)
		f.NeedPadding()
	}

	f.FormatDirectiveList(field.Directives)

	f.FormatSelectionSet(field.SelectionSet)
}

func (f *Formatter) FormatFragmentSpread(spread *ast.FragmentSpread) {
	f.WriteWord("...").WriteWord(spread.Name)

	f.FormatDirectiveList(spread.Directives)
}

func (f *Formatter) FormatInlineFragment(inline *ast.InlineFragment) {
	f.WriteWord("...")
	if inline.TypeCondition != "" {
		f.WriteWord("on").WriteWord(inline.TypeCondition)
	}

	f.FormatDirectiveList(inline.Directives)

	f.FormatSelectionSet(inline.SelectionSet)
}

func (f *Formatter) FormatType(t *ast.Type) {
	f.WriteWord(t.String())
}

func (f *Formatter) FormatValue(value *ast.Value) {
	f.WriteString(value.String())
}
