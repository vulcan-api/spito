package api

type InfoInterface interface {
	Log(...string)
	Debug(...string)
	Error(...string)
	Warn(...string)
	Important(...string)
}
