package guiApi

import (
	"github.com/godbus/dbus"
	"strings"
)

type InfoApi struct {
	BusObject dbus.BusObject
}

func (i InfoApi) Log(args ...string) {
	_ = sendToDbusMethod(i.BusObject, "log", args...)
}

func (i InfoApi) Debug(args ...string) {
	_ = sendToDbusMethod(i.BusObject, "debug", args...)
}

func (i InfoApi) Error(args ...string) {
	_ = sendToDbusMethod(i.BusObject, "error", args...)
}

func (i InfoApi) Warn(args ...string) {
	_ = sendToDbusMethod(i.BusObject, "warn", args...)
}

func (i InfoApi) Important(args ...string) {
	_ = sendToDbusMethod(i.BusObject, "important", args...)
}

// Most of the time we ignore potential error because it is not really important
// and our app can work even if error is thrown
func sendToDbusMethod(busObject dbus.BusObject, logType string, values ...string) error {
	mergedValues := strings.Join(values[:], "")
	call := busObject.Call("Echo", 0, logType, mergedValues)
	return call.Err
}
