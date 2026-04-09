package dto

type OTPPurpose string

const (
	EmailVerification OTPPurpose = "EMAIL_VERIFICATION"
	PasswordReset     OTPPurpose = "PASSWORD_RESET"
	Login             OTPPurpose = "LOGIN"
)

// String returns string representation of [OTPPurpose]
func (o OTPPurpose) String() string {
	switch o {
	case EmailVerification:
		return "EMAIL_VERIFICATION"
	case PasswordReset:
		return "PASSWORD_RESET"
	case Login:
		return "LOGIN"
	default:
		return "invalid enum type"
	}
}

// IsValid returns whether the value of [OTPPurpose] is valid.
func (o OTPPurpose) IsValid() bool {
	switch o {
	case EmailVerification, PasswordReset, Login:
		return true
	default:
		return false
	}
}

type EmailDeliveryType string

const (
	TypeVerificationEmail  EmailDeliveryType = "email:account-verification-otp"
	TypeLoginEmail         EmailDeliveryType = "email:login-otp"
	TypePasswordResetEmail EmailDeliveryType = "email:password-reset-otp"
	TypeSecurityAlertEmail EmailDeliveryType = "email:security-alert"
)

// IsValid returns whether the value of [EmailDeliveryType] is valid.
func (e EmailDeliveryType) IsValid() bool {
	switch e {
	case TypeVerificationEmail, TypeLoginEmail, TypePasswordResetEmail, TypeSecurityAlertEmail:
		return true
	default:
		return false
	}
}

type OAuthProvider string

const (
	OAuthProviderGoogle OAuthProvider = "google"
	OAuthProviderGithub OAuthProvider = "github"
)

// String returns string representation of [OAuthProvider]
func (o OAuthProvider) String() string {
	switch o {
	case OAuthProviderGoogle:
		return "google"
	case OAuthProviderGithub:
		return "github"
	default:
		return ""
	}
}
