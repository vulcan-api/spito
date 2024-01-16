package vrct

type VRCT interface {
	InnerValidate() error
	Apply() error
	DeleteRuntimeTemp() error
	Revert() error

	//Serialize() ([]byte, error)
	//Deserialize([]byte) (VRCT, error)
}
