package ipcheck

// Checker resepresents an IP checking service.
type Checker interface {
	GetIP() (string, error)
}
