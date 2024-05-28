package completion

type ItemKind uint

const (
	Text ItemKind = iota + 1
	Method
	Function
	Constructor
	Field
	Variable
	Class
	Interface
	Module
	Property
	Unit
	Value
	Enum
	Keyword
	Snippet
	Color
	File
	Reference
	Folder
	EnumMember
	Constant
	Struct
	Event
	Operator
	TypeParameter
)

type TriggerKind uint

const (
	Invoked TriggerKind = iota + 1
	TriggerCharacter
	TriggerForIncompleteCompletions
)
