package shared

type InfoInterface interface {
	Log(...any)
	Debug(...any)
	Error(...any)
	Warn(...any)
	Important(...any)
}
