package v1

// AuthToken is a JSON Web Token.
//
// All fields related to times are stored as UTC epochs in seconds.
type AuthToken interface {

	// GetSubject is the intended principal of the token.
	GetSubject() string

	// GetExpires is the time at which the token expires.
	GetExpires() int64

	// GetNotBefore is the the time at which the token becomes valid.
	GetNotBefore() int64

	// GetIssuedAt is the time at which the token was issued.
	GetIssuedAt() int64

	// GetEncoded is the encoded JWT string.
	GetEncoded() string
}

// AuthConfig is the auth configuration.
type AuthConfig interface {

	// GetDisabled is a flag indicating whether the auth configuration is
	// disabled.
	GetDisabled() bool

	// GetAllow is a list of allowed tokens.
	GetAllow() []string

	// GetDeny is a list of denied tokens.
	GetDeny() []string

	// GetKey is the signing key.
	GetKey() []byte

	// GetAlg is the cryptographic algorithm used to sign and verify the token.
	GetAlg() string
}
