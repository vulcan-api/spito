package shared

import (
	"fmt"
	"github.com/godbus/dbus/v5"
	"os"
)

func DBusInterfaceId() string {
	return os.Getenv("DBUS_INTERFACE_ID")
}

func DBusMethodName(name string) string {
	return fmt.Sprint(os.Getenv("DBUS_INTERFACE_ID"), ".", name)
}

func DBusMethod(conn *dbus.Conn, methodName string, args ...any) *dbus.Call {
	object := DBusObject(conn)
	return object.Call(DBusMethodName(methodName), 0, args...)
}

func DBusMethodP(conn *dbus.Conn, methodName, panicMessage string, args ...any) {
	call := DBusMethod(conn, methodName, args...)
	if call.Err != nil {
		panic(panicMessage + ": " + call.Err.Error())
	}
}

func DBusObjectPath() dbus.ObjectPath {
	return dbus.ObjectPath(os.Getenv("DBUS_OBJECT_PATH"))
}

func DBusObject(conn *dbus.Conn) dbus.BusObject {
	dbusId := DBusInterfaceId()
	dbusPath := DBusObjectPath()

	return conn.Object(dbusId, dbusPath)
}
