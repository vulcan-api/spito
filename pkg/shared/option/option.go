package option

type Type uint

const (
	Int Type = iota
	UInt
	Float
	String
	Bool
	// Array
	// Struct
	// Enum
	Any
	Unknown
)

type Option struct {
	Name         string
	DefaultValue any
	Type         Type
}
