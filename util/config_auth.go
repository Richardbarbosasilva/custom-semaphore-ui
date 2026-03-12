package util

type RecaptchaConfig struct {
	Enabled string `json:"enabled,omitempty" env:"SEMAPHORE_RECAPTCHA_ENABLED"`
	SiteKey string `json:"site_key,omitempty" env:"SEMAPHORE_RECAPTCHA_SITE_KEY"`
}

type EmailAuthConfig struct {
	Enabled                  bool     `json:"enabled" env:"SEMAPHORE_EMAIL_2TP_ENABLED"`
	AllowLoginAsExternalUser bool     `json:"allow_login_as_external_user" env:"SEMAPHORE_EMAIL_2TP_ALLOW_LOGIN_AS_EXTERNAL_USER"`
	AllowCreateExternalUsers bool     `json:"allow_create_external_user" env:"SEMAPHORE_EMAIL_2TP_ALLOW_CREATE_EXTERNAL_USER"`
	AllowedDomains           []string `json:"allowed_domains" env:"SEMAPHORE_EMAIL_2TP_ALLOWED_DOMAINS"`
	DisableForOidc           bool     `json:"disable_for_oidc" env:"SEMAPHORE_EMAIL_2TP_DISABLE_FOR_OIDC"`
}

type AuthConfig struct {
	Totp  *TotpConfig      `json:"totp,omitempty"`
	Email *EmailAuthConfig `json:"email,omitempty"`
}

func GetAuthConfig() *AuthConfig {
	if Config.Auth == nil {
		Config.Auth = &AuthConfig{}
	}

	return Config.Auth
}

func GetTotpConfig() *TotpConfig {
	auth := GetAuthConfig()

	if auth.Totp == nil {
		auth.Totp = &TotpConfig{}
	}

	return auth.Totp
}

func GetEmailAuthConfig() *EmailAuthConfig {
	auth := GetAuthConfig()

	if auth.Email == nil {
		auth.Email = &EmailAuthConfig{}
	}

	return auth.Email
}

func GetLdapMappings() *LdapMappings {
	if Config.LdapMappings == nil {
		Config.LdapMappings = &LdapMappings{}
	}

	if Config.LdapMappings.DN == "" {
		Config.LdapMappings.DN = "dn"
	}

	if Config.LdapMappings.Mail == "" {
		Config.LdapMappings.Mail = "mail"
	}

	if Config.LdapMappings.UID == "" {
		Config.LdapMappings.UID = "uid"
	}

	if Config.LdapMappings.CN == "" {
		Config.LdapMappings.CN = "cn"
	}

	return Config.LdapMappings
}
