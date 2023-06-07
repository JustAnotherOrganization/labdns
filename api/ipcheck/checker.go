package ipcheck

type Checker interface {
	GetIP() (string, error)
}
