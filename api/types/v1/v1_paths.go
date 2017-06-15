package v1

// PathConfig contains the path configuration for the application.
type PathConfig interface {

	// GetToken returns the app token.
	GetToken() string

	// GetHome returns the path to the system, root, data directory.
	GetHome() string

	// GetEtc returns the path to the etc directory.
	GetEtc() string

	// GetLib returns the path to the lib directory.
	GetLib() string

	// GetMod returns the path to the mod directory.
	GetMod() string

	// GetLog returns the path to the log directory.
	GetLog() string

	// GetRun returns the path to the run directory.
	GetRun() string

	// GetTLS returns the path to the tls directory.
	GetTLS() string

	// GetLSX returns the path to the executor.
	GetLSX() string

	// GetDefaultTLSCertFile returns the path to the default TLS cert
	// file.
	GetDefaultTLSCertFile() string

	// GetDefaultTLSKeyFile returns the path to the default TLS key
	// file.
	GetDefaultTLSKeyFile() string

	// GetDefaultTLSTrustedRootsFile returns the path to the default
	// TLS trusted roots file.
	GetDefaultTLSTrustedRootsFile() string

	// GetDefaultTLSKnownHosts returns the default path to the TLS
	// known hosts file.
	GetDefaultTLSKnownHosts() string

	// GetUserHome returns the path to the user, root, data directory.
	GetUserHome() string

	// GetUserDefaultTLSKnownHosts returns the default path to the
	// user's TLS known hosts file.
	GetUserDefaultTLSKnownHosts() string
}
