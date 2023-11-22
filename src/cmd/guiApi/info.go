package guiApi

import "github.com/godbus/dbus"

const (
	dbusServiceName   = "org.spito.gui"
	dbusObjectPath    = "/org/spito/gui"
	dbusInterfaceName = "org.spito.gui"
)

type InfoApi struct {
	BusObject dbus.BusObject
}

func (i InfoApi) Log(args ...string) {
	_ = sendToDbusMethod(i.BusObject, "Echo", args...)
}

func (i InfoApi) Debug(args ...string) {
	_ = sendToDbusMethod(i.BusObject, "Echo", args...)
}

func (i InfoApi) Error(args ...string) {
	_ = sendToDbusMethod(i.BusObject, "Echo", args...)
}

func (i InfoApi) Warn(args ...string) {
	_ = sendToDbusMethod(i.BusObject, "Echo", args...)
}

func (i InfoApi) Important(args ...string) {
	_ = sendToDbusMethod(i.BusObject, "Echo", args...)
}

// Most of the time we ignore potential error because it is not really important
// and our app can work even if error is thrown
func sendToDbusMethod(busObject dbus.BusObject, methodName string, values ...string) error {
	_values := intoInterfaceArray(values)
	call := busObject.Call(dbusInterfaceName+"."+methodName, 0, _values)
	return call.Err
}

func intoInterfaceArray(args []string) []any {
	result := make([]any, len(args))

	for i, e := range args {
		result[i+1] = e
	}

	return result
}
