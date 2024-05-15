package indexables

import (
	"fmt"

	"github.com/pherrymason/c3-lsp/option"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type Struct struct {
	members    []StructMember
	isUnion    bool
	implements []string
	BaseIndexable
}

func NewStruct(name string, interfaces []string, members []StructMember, module string, docId string, idRange Range, docRange Range) Struct {
	return Struct{
		members:    members,
		isUnion:    false,
		implements: interfaces,
		BaseIndexable: NewBaseIndexable(
			name,
			module,
			docId,
			idRange,
			docRange,
			protocol.CompletionItemKindStruct,
		),
	}
}

func NewUnion(name string, members []StructMember, module string, docId string, idRange Range, docRange Range) Struct {
	return Struct{
		members: members,
		isUnion: true,
		BaseIndexable: NewBaseIndexable(
			name,
			module,
			docId,
			idRange,
			docRange,
			protocol.CompletionItemKindStruct,
		),
	}
}

func (s Struct) GetMembers() []StructMember {
	return s.members
}

func (s Struct) GetInterfaces() []string {
	return s.implements
}

func (s Struct) IsUnion() bool {
	return s.isUnion
}

func (s Struct) GetHoverInfo() string {
	return fmt.Sprintf("%s", s.name)
}

type StructMember struct {
	baseType Type
	bitRange option.Option[[2]uint]
	BaseIndexable
}

func (m StructMember) GetType() Type {
	return m.baseType
}

func (m StructMember) GetBitRange() [2]uint {
	return m.bitRange.Get()
}

func (s StructMember) GetHoverInfo() string {
	return fmt.Sprintf("%s %s", s.baseType, s.name)
}

func NewStructMember(name string, fieldType string, bitRanges option.Option[[2]uint], module string, docId string, idRange Range) StructMember {
	return StructMember{
		baseType: NewTypeFromString(fieldType),
		bitRange: bitRanges,
		BaseIndexable: NewBaseIndexable(
			name,
			module,
			docId,
			idRange,
			NewRange(0, 0, 0, 0),
			protocol.CompletionItemKindField,
		),
	}
}
