package collectionx

type jsonStringer interface {
	ToJSON() ([]byte, error)
	String() string
}

type sized interface {
	Len() int
	IsEmpty() bool
}

type clearable interface {
	Clear()
}

type clonable[T any] interface {
	Clone() T
}

type snapshotable[T any] interface {
	Snapshot() T
}
