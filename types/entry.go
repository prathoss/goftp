package types

const (
	TypeFile = iota
	TypeDirectory
	TypeLink
)

type Entry struct {
	Name string
	Type int
	Size uint64
}

func (e Entry) TypeString() string {
	switch e.Type {
	case TypeFile:
		return "f"
	case TypeDirectory:
		return "d"
	default:
		return "l"
	}
}
