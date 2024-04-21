package parser

import (
	"github.com/pherrymason/c3-lsp/lsp/document"
	idx "github.com/pherrymason/c3-lsp/lsp/indexables"
	sitter "github.com/smacker/go-tree-sitter"
)

/*
struct_declaration: $ => seq(

	$._struct_or_union,
	field('name', $.type_ident),
	optional($.interface_impl),
	optional($.attributes),
	field('body', $.struct_body),

),
_struct_or_union: _ => choice('struct', 'union'),
struct_body: $ => seq(

	  '{',
	  // NOTE Allowing empty struct to not be too strict.
	  repeat($.struct_member_declaration),
	  '}',
	),

struct_member_declaration: $ => choice(

	  seq(field('type', $.type), $.identifier_list, optional($.attributes), ';'),
	  seq($._struct_or_union, optional($.ident), optional($.attributes), field('body', $.struct_body)),
	  seq('bitstruct', optional($.ident), ':', $.type, optional($.attributes), field('body', $.bitstruct_body)),
	  seq('inline', field('type', $.type), optional($.ident), optional($.attributes), ';'),
	),
*/
func (p *Parser) nodeToStruct(doc *document.Document, node *sitter.Node, sourceCode []byte) idx.Struct {
	nameNode := node.ChildByFieldName("name")
	name := nameNode.Content(sourceCode)
	isUnion := false

	for i := uint32(0); i < node.ChildCount(); i++ {
		child := node.Child(int(i))

		switch child.Type() {
		case "union":
			isUnion = true
		case "interface_impl":
			// TODO
		case "attributes":
			// TODO attributes
		}
	}

	// TODO parse attributes
	bodyNode := node.ChildByFieldName("body")
	structFields := make([]idx.StructMember, 0)

	for i := uint32(0); i < bodyNode.ChildCount(); i++ {
		memberNode := bodyNode.Child(int(i))
		//fmt.Println("body child:", memberNode.Type())
		if memberNode.Type() != "struct_member_declaration" {
			continue
		}

		var fieldType string
		var identifiers []string
		var identifiersRange []idx.Range

		for x := uint32(0); x < memberNode.ChildCount(); x++ {
			n := memberNode.Child(int(x))
			switch n.Type() {
			case "type":
				fieldType = n.Content(sourceCode)
			case "identifier_list":
				for j := uint32(0); j < n.ChildCount(); j++ {
					identifiers = append(identifiers, n.Child(int(j)).Content(sourceCode))
					identifiersRange = append(identifiersRange,
						idx.NewRangeFromSitterPositions(n.StartPoint(), n.EndPoint()),
					)
				}
			case "attributes":
				// TODO
			}
		}

		for y := 0; y < len(identifiers); y++ {
			structFields = append(
				structFields,
				idx.NewStructMember(
					identifiers[y],
					fieldType,
					identifiersRange[y]),
			)
		}
	}

	var _struct idx.Struct
	if isUnion {
		_struct = idx.NewUnion(
			name,
			structFields,
			doc.ModuleName,
			doc.URI,
			idx.NewRangeFromSitterPositions(nameNode.StartPoint(), nameNode.EndPoint()),
			idx.NewRangeFromSitterPositions(node.StartPoint(), node.EndPoint()),
		)
	} else {
		_struct = idx.NewStruct(
			name,
			structFields,
			doc.ModuleName,
			doc.URI,
			idx.NewRangeFromSitterPositions(nameNode.StartPoint(), nameNode.EndPoint()),
			idx.NewRangeFromSitterPositions(node.StartPoint(), node.EndPoint()),
		)
	}

	return _struct
}
