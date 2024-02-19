package option

type Type uint

const (
	Any Type = iota
	Int
	UInt
	Float
	String
	Bool
	// Array
	// Struct
	// Enum
	Unknown
)

type Option struct {
	Name           string
	DefaultValue   any
	Type           Type
	Optional       bool
	PossibleValues []any
}

func GetType(rawValue any) Type {
	_, ok := rawValue.(int)
	if ok {
		return Int
	}
	_, ok = rawValue.(uint)
	if ok {
		return UInt
	}
	_, ok = rawValue.(float64)
	if ok {
		return Float
	}
	_, ok = rawValue.(bool)
	if ok {
		return Bool
	}
	_, ok = rawValue.(string)
	if ok {
		return String
	}
	return Unknown
}
