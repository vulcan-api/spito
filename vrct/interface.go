package vrct

type VRCT interface {
	InnerValidate() error
	Apply() error

	//TODO: implement soon
	//Revert() error

	// Currently unimportant
	//Serialize() []byte
	//Deserialize([]byte) (VRCT, error)
}
