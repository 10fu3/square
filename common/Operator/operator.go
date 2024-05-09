package Operator

type Type uint

const (
	True   Type = iota
	Eq     Type = iota
	Gt     Type = iota
	Gte    Type = iota
	Lt     Type = iota
	Lte    Type = iota
	Neq    Type = iota
	In     Type = iota
	NotIn  Type = iota
	IsNull Type = iota
)

func (o Type) String() string {
	switch o {
	case Eq:
		return "="
	case Gt:
		return ">"
	case Gte:
		return ">="
	case Lt:
		return "<"
	case Lte:
		return "<="
	case Neq:
		return "!="
	case In:
		return "IN"
	case NotIn:
		return "NOT IN"
	case IsNull:
		return "IS NULL"
	default:
		return "1 = 1"
	}
}
