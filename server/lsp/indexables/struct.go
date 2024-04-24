package indexables

import (
	"fmt"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

type Struct struct {
	name    string
	members []StructMember
	isUnion bool
	BaseIndexable
}

func NewStruct(name string, members []StructMember, module string, docId string, idRange Range, docRange Range) Struct {
	return Struct{
		name:    name,
		members: members,
		isUnion: false,
		BaseIndexable: BaseIndexable{
			module:      module,
			documentURI: docId,
			idRange:     idRange,
			docRange:    docRange,
			Kind:        protocol.CompletionItemKindStruct,
		},
	}
}

func NewUnion(name string, members []StructMember, module string, docId string, idRange Range, docRange Range) Struct {
	return Struct{
		name:    name,
		members: members,
		isUnion: true,
		BaseIndexable: BaseIndexable{
			module:      module,
			documentURI: docId,
			idRange:     idRange,
			docRange:    docRange,
			Kind:        protocol.CompletionItemKindStruct,
		},
	}
}

func (s Struct) GetName() string {
	return s.name
}

func (s Struct) GetMembers() []StructMember {
	return s.members
}

func (s Struct) GetModule() string {
	return s.module
}

func (s Struct) GetKind() protocol.CompletionItemKind {
	return s.Kind
}

func (s Struct) IsUnion() bool {
	return s.isUnion
}

func (s Struct) GetDocumentURI() string {
	return s.documentURI
}

func (s Struct) GetIdRange() Range {
	return s.idRange
}
func (s Struct) GetDocumentRange() Range {
	return s.docRange
}

func (s Struct) GetHoverInfo() string {
	return fmt.Sprintf("%s", s.name)
}

type StructMember struct {
	name     string
	baseType string
	BaseIndexable
}

func (m StructMember) GetName() string {
	return m.name
}

func (m StructMember) GetType() string {
	return m.baseType
}

func (m StructMember) GetIdRange() Range {
	return m.idRange
}

func (m StructMember) GetDocumentRange() Range {
	return m.docRange
}

func (m StructMember) GetDocumentURI() string {
	return m.documentURI
}

func (s StructMember) GetHoverInfo() string {
	return fmt.Sprintf("%s", s.name)
}
func (s StructMember) GetKind() protocol.CompletionItemKind {
	return s.Kind
}
func (s StructMember) GetModule() string {
	return s.module
}

func NewStructMember(name string, baseType string, posRange Range) StructMember {
	return StructMember{
		name:     name,
		baseType: baseType,
		BaseIndexable: BaseIndexable{
			idRange: posRange,
		},
	}
}
