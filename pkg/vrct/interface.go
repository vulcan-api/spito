package vrct

type VRCT interface {
	InnerValidate() error
	Apply() error
	DeleteRuntimeTemp() error
	Revert() error

	//TODO: implement soon:
	//Serialize() ([]byte, error)
	//Deserialize([]byte) (VRCT, error)
}
