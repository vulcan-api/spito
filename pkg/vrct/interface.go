package vrct

type VRCT interface {
	Apply() (int, error)
	DeleteRuntimeTemp() error
	Revert() error

	//Serialize() ([]byte, error)
	//Deserialize([]byte) (VRCT, error)
}
