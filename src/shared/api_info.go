package shared

type InfoInterface interface {
	Log(...string)
	Debug(...string)
	Error(...string)
	Warn(...string)
	Important(...string)
}
