package domain

// Sensitivity represents the security classification of data collected by scanners.
type Sensitivity int

const (
	// Public indicates no security concerns (e.g., Homebrew packages, VSCode extensions).
	Public Sensitivity = iota
	// Sensitive indicates data that may contain private info (e.g., .npmrc, AWS config).
	Sensitive
	// Secret indicates credentials, keys, or tokens (e.g., SSH keys, .env files).
	Secret
)

// String returns the lowercase string representation of the Sensitivity level.
func (s Sensitivity) String() string {
	switch s {
	case Public:
		return "public"
	case Sensitive:
		return "sensitive"
	case Secret:
		return "secret"
	default:
		return "unknown"
	}
}
