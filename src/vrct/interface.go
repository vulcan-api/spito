package vrct

import "fmt"

type VRCT interface {
	InnerValidate() error // TODO change type into *ConflictError
	Apply() error

	//TODO: implement soon
	//Revert() error

	// Currently unimportant
	//IsApplied() bool
	//DoesMadeThisChange(VRCT) bool
	//Serialize() []byte
	//Deserialize([]byte) (VRCT, error)
}

type ConflictError struct {
	Err                error
	Element            VRCT
	ElementConflicting VRCT
}

func (c ConflictError) Error() string {
	return fmt.Sprintf("element are in conflict!\n %v", c.Err)
}
