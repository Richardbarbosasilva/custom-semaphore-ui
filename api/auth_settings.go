package api

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
)

type ldapRuntimeConfig struct {
	Enabled      bool
	Server       string
	NeedTLS      bool
	BindDN       string
	BindPassword string
	SearchDN     string
	SearchFilter string
	Mappings     util.LdapMappings
}

type authSettingsResponse struct {
	LDAP ldapSettingsResponse `json:"ldap"`
	Totp totpSettingsResponse `json:"totp"`
}

type ldapSettingsResponse struct {
	Enabled         bool                 `json:"enabled"`
	Server          string               `json:"server"`
	NeedTLS         bool                 `json:"need_tls"`
	BindDN          string               `json:"bind_dn"`
	HasBindPassword bool                 `json:"has_bind_password"`
	SearchDN        string               `json:"search_dn"`
	SearchFilter    string               `json:"search_filter"`
	Mappings        ldapMappingsResponse `json:"mappings"`
}

type ldapMappingsResponse struct {
	DN   string `json:"dn"`
	UID  string `json:"uid"`
	CN   string `json:"cn"`
	Mail string `json:"mail"`
}

type totpSettingsResponse struct {
	Enabled       bool   `json:"enabled"`
	AllowRecovery bool   `json:"allow_recovery"`
	Issuer        string `json:"issuer"`
}

type authSettingsRequest struct {
	LDAP ldapSettingsRequest `json:"ldap"`
	Totp totpSettingsRequest `json:"totp"`
}

type ldapSettingsRequest struct {
	Enabled           bool                `json:"enabled"`
	Server            string              `json:"server"`
	NeedTLS           bool                `json:"need_tls"`
	BindDN            string              `json:"bind_dn"`
	BindPassword      string              `json:"bind_password"`
	ClearBindPassword bool                `json:"clear_bind_password"`
	SearchDN          string              `json:"search_dn"`
	SearchFilter      string              `json:"search_filter"`
	Mappings          ldapMappingsRequest `json:"mappings"`
}

type ldapMappingsRequest struct {
	DN   string `json:"dn"`
	UID  string `json:"uid"`
	CN   string `json:"cn"`
	Mail string `json:"mail"`
}

type totpSettingsRequest struct {
	Enabled       bool   `json:"enabled"`
	AllowRecovery bool   `json:"allow_recovery"`
	Issuer        string `json:"issuer"`
}

type ldapTestRequest struct {
	LDAP     ldapSettingsRequest `json:"ldap"`
	Login    string              `json:"login"`
	Password string              `json:"password"`
}

type ldapTestResponse struct {
	Message string               `json:"message"`
	User    *ldapTestUserPreview `json:"user,omitempty"`
}

type ldapTestUserPreview struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

const (
	defaultLDAPDNMapping     = "dn"
	defaultLDAPMailMapping   = "mail"
	defaultLDAPUIDMapping    = "uid"
	defaultLDAPCNMapping     = "cn"
	defaultADDNMapping       = "distinguishedName"
	defaultADMailMapping     = "mail"
	defaultADUIDMapping      = "sAMAccountName"
	defaultADCNMapping       = "displayName"
	defaultADSearchFilter    = "(&(objectClass=user)(sAMAccountName=%s))"
	defaultGenericLDAPFilter = "(uid=%s)"
)

func looksLikeActiveDirectory(filter string) bool {
	return strings.Contains(strings.ToLower(strings.TrimSpace(filter)), "samaccountname")
}

func defaultLDAPMappingsForFilter(filter string) util.LdapMappings {
	if looksLikeActiveDirectory(filter) {
		return util.LdapMappings{
			DN:   defaultADDNMapping,
			Mail: defaultADMailMapping,
			UID:  defaultADUIDMapping,
			CN:   defaultADCNMapping,
		}
	}

	return util.LdapMappings{
		DN:   defaultLDAPDNMapping,
		Mail: defaultLDAPMailMapping,
		UID:  defaultLDAPUIDMapping,
		CN:   defaultLDAPCNMapping,
	}
}

func normalizeLDAPMappingValue(value string, fallback string, useFallbackFor string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" || strings.EqualFold(trimmedValue, useFallbackFor) {
		return fallback
	}

	return trimmedValue
}

func getDefaultLdapMappings() ldapMappingsResponse {
	normalized := normalizeLDAPConfig(getCurrentLDAPRuntimeConfig())
	mappings := normalized.Mappings

	return ldapMappingsResponse{
		DN:   mappings.DN,
		UID:  mappings.UID,
		CN:   mappings.CN,
		Mail: mappings.Mail,
	}
}

func getCurrentAuthSettings() authSettingsResponse {
	ldapCfg := normalizeLDAPConfig(getCurrentLDAPRuntimeConfig())
	totp := util.GetTotpConfig()

	return authSettingsResponse{
		LDAP: ldapSettingsResponse{
			Enabled:         ldapCfg.Enabled,
			Server:          ldapCfg.Server,
			NeedTLS:         ldapCfg.NeedTLS,
			BindDN:          ldapCfg.BindDN,
			HasBindPassword: util.Config.LdapBindPassword != "",
			SearchDN:        ldapCfg.SearchDN,
			SearchFilter:    ldapCfg.SearchFilter,
			Mappings:        getDefaultLdapMappings(),
		},
		Totp: totpSettingsResponse{
			Enabled:       totp.Enabled,
			AllowRecovery: totp.AllowRecovery,
			Issuer:        totp.Issuer,
		},
	}
}

func ldapRequestToRuntimeConfig(req ldapSettingsRequest, currentPassword string) ldapRuntimeConfig {
	bindPassword := currentPassword

	switch {
	case req.ClearBindPassword:
		bindPassword = ""
	case req.BindPassword != "":
		bindPassword = req.BindPassword
	}

	return ldapRuntimeConfig{
		Enabled:      req.Enabled,
		Server:       strings.TrimSpace(req.Server),
		NeedTLS:      req.NeedTLS,
		BindDN:       strings.TrimSpace(req.BindDN),
		BindPassword: bindPassword,
		SearchDN:     strings.TrimSpace(req.SearchDN),
		SearchFilter: strings.TrimSpace(req.SearchFilter),
		Mappings: util.LdapMappings{
			DN:   strings.TrimSpace(req.Mappings.DN),
			UID:  strings.TrimSpace(req.Mappings.UID),
			CN:   strings.TrimSpace(req.Mappings.CN),
			Mail: strings.TrimSpace(req.Mappings.Mail),
		},
	}
}

func getCurrentLDAPRuntimeConfig() ldapRuntimeConfig {
	mappings := util.GetLdapMappings()

	return ldapRuntimeConfig{
		Enabled:      util.Config.LdapEnable,
		Server:       util.Config.LdapServer,
		NeedTLS:      util.Config.LdapNeedTLS,
		BindDN:       util.Config.LdapBindDN,
		BindPassword: util.Config.LdapBindPassword,
		SearchDN:     util.Config.LdapSearchDN,
		SearchFilter: util.Config.LdapSearchFilter,
		Mappings: util.LdapMappings{
			DN:   mappings.DN,
			UID:  mappings.UID,
			CN:   mappings.CN,
			Mail: mappings.Mail,
		},
	}
}

func normalizeLDAPConfig(cfg ldapRuntimeConfig) ldapRuntimeConfig {
	defaultMappings := defaultLDAPMappingsForFilter(cfg.SearchFilter)

	if cfg.SearchFilter == "" {
		if looksLikeActiveDirectory(cfg.SearchFilter) {
			cfg.SearchFilter = defaultADSearchFilter
		} else {
			cfg.SearchFilter = defaultGenericLDAPFilter
		}
	}

	cfg.Mappings.DN = normalizeLDAPMappingValue(cfg.Mappings.DN, defaultMappings.DN, defaultLDAPDNMapping)
	cfg.Mappings.UID = normalizeLDAPMappingValue(cfg.Mappings.UID, defaultMappings.UID, defaultLDAPUIDMapping)
	cfg.Mappings.CN = normalizeLDAPMappingValue(cfg.Mappings.CN, defaultMappings.CN, defaultLDAPCNMapping)
	cfg.Mappings.Mail = normalizeLDAPMappingValue(cfg.Mappings.Mail, defaultMappings.Mail, defaultLDAPMailMapping)

	return cfg
}

func validateLDAPConfig(cfg ldapRuntimeConfig) error {
	if strings.TrimSpace(cfg.Server) == "" {
		return fmt.Errorf("LDAP server is required")
	}

	if strings.TrimSpace(cfg.SearchDN) == "" {
		return fmt.Errorf("LDAP search DN is required")
	}

	return nil
}

func dialLDAP(cfg ldapRuntimeConfig) (*ldap.Conn, error) {
	cfg = normalizeLDAPConfig(cfg)

	var (
		conn *ldap.Conn
		err  error
	)

	if cfg.NeedTLS {
		conn, err = ldap.DialTLS("tcp", cfg.Server, &tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		conn, err = ldap.Dial("tcp", cfg.Server)
	}

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func bindLDAPReader(conn *ldap.Conn, cfg ldapRuntimeConfig) error {
	if cfg.BindDN == "" {
		return nil
	}

	return conn.Bind(cfg.BindDN, cfg.BindPassword)
}

func searchLDAPUser(conn *ldap.Conn, cfg ldapRuntimeConfig, username string, attrs []string) (*ldap.SearchResult, error) {
	filter := fmt.Sprintf(cfg.SearchFilter, ldap.EscapeFilter(strings.TrimSpace(username)))

	return conn.Search(ldap.NewSearchRequest(
		cfg.SearchDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		attrs,
		nil,
	))
}

func findLDAPUserWithConfig(cfg ldapRuntimeConfig, username string, password string) (*db.User, error) {
	cfg = normalizeLDAPConfig(cfg)

	if !cfg.Enabled {
		return nil, fmt.Errorf("LDAP not configured")
	}

	if err := validateLDAPConfig(cfg); err != nil {
		return nil, err
	}

	conn, err := dialLDAP(cfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close() //nolint:errcheck

	if err = bindLDAPReader(conn, cfg); err != nil {
		return nil, err
	}

	searchResult, err := searchLDAPUser(conn, cfg, username, []string{cfg.Mappings.DN})
	if err != nil {
		return nil, err
	}

	if len(searchResult.Entries) < 1 {
		return nil, nil
	}

	if len(searchResult.Entries) > 1 {
		return nil, fmt.Errorf("too many entries returned")
	}

	userDN := searchResult.Entries[0].DN
	if err = conn.Bind(userDN, password); err != nil {
		return nil, err
	}

	if err = bindLDAPReader(conn, cfg); err != nil {
		return nil, err
	}

	searchResult, err = searchLDAPUser(conn, cfg, username, []string{
		cfg.Mappings.DN,
		cfg.Mappings.Mail,
		cfg.Mappings.UID,
		cfg.Mappings.CN,
	})
	if err != nil {
		return nil, err
	}

	if len(searchResult.Entries) <= 0 {
		return nil, fmt.Errorf("ldap search returned no entries")
	}

	entry := convertEntryToMap(searchResult.Entries[0])

	prepareClaims(entry)

	claims, err := parseLDAPClaims(entry, &cfg.Mappings, username)
	if err != nil {
		return nil, err
	}

	ldapUser := db.User{
		Username: strings.ToLower(claims.username),
		Created:  tz.Now(),
		Name:     claims.name,
		Email:    claims.email,
		External: true,
		Alert:    false,
	}

	err = db.ValidateUser(ldapUser)
	if err != nil {
		return nil, err
	}

	return &ldapUser, nil
}

func testLDAPReaderConnection(cfg ldapRuntimeConfig) error {
	cfg = normalizeLDAPConfig(cfg)

	if err := validateLDAPConfig(cfg); err != nil {
		return err
	}

	conn, err := dialLDAP(cfg)
	if err != nil {
		return err
	}
	defer conn.Close() //nolint:errcheck

	return bindLDAPReader(conn, cfg)
}

func persistAuthSettings(store db.Store, settings authSettingsRequest, bindPassword string) error {
	options := map[string]string{
		"ldap_enable":              strconv.FormatBool(settings.LDAP.Enabled),
		"ldap_server":              strings.TrimSpace(settings.LDAP.Server),
		"ldap_needtls":             strconv.FormatBool(settings.LDAP.NeedTLS),
		"ldap_binddn":              strings.TrimSpace(settings.LDAP.BindDN),
		"ldap_bindpassword":        bindPassword,
		"ldap_searchdn":            strings.TrimSpace(settings.LDAP.SearchDN),
		"ldap_searchfilter":        strings.TrimSpace(settings.LDAP.SearchFilter),
		"ldap_mappings.dn":         strings.TrimSpace(settings.LDAP.Mappings.DN),
		"ldap_mappings.uid":        strings.TrimSpace(settings.LDAP.Mappings.UID),
		"ldap_mappings.cn":         strings.TrimSpace(settings.LDAP.Mappings.CN),
		"ldap_mappings.mail":       strings.TrimSpace(settings.LDAP.Mappings.Mail),
		"auth.totp.enabled":        strconv.FormatBool(settings.Totp.Enabled),
		"auth.totp.allow_recovery": strconv.FormatBool(settings.Totp.AllowRecovery),
		"auth.totp.app_name":       strings.TrimSpace(settings.Totp.Issuer),
	}

	for key, value := range options {
		if err := store.SetOption(key, value); err != nil {
			return err
		}
	}

	return nil
}

func applyAuthSettings(settings authSettingsRequest, bindPassword string) {
	util.Config.LdapEnable = settings.LDAP.Enabled
	util.Config.LdapServer = strings.TrimSpace(settings.LDAP.Server)
	util.Config.LdapNeedTLS = settings.LDAP.NeedTLS
	util.Config.LdapBindDN = strings.TrimSpace(settings.LDAP.BindDN)
	util.Config.LdapBindPassword = bindPassword
	util.Config.LdapSearchDN = strings.TrimSpace(settings.LDAP.SearchDN)
	util.Config.LdapSearchFilter = strings.TrimSpace(settings.LDAP.SearchFilter)

	mappings := util.GetLdapMappings()
	mappings.DN = strings.TrimSpace(settings.LDAP.Mappings.DN)
	mappings.UID = strings.TrimSpace(settings.LDAP.Mappings.UID)
	mappings.CN = strings.TrimSpace(settings.LDAP.Mappings.CN)
	mappings.Mail = strings.TrimSpace(settings.LDAP.Mappings.Mail)

	totp := util.GetTotpConfig()
	totp.Enabled = settings.Totp.Enabled
	totp.AllowRecovery = settings.Totp.AllowRecovery
	totp.Issuer = strings.TrimSpace(settings.Totp.Issuer)
}

func getAuthSettings(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, getCurrentAuthSettings())
}

func setAuthSettings(w http.ResponseWriter, r *http.Request) {
	var settings authSettingsRequest
	if !helpers.Bind(w, r, &settings) {
		return
	}

	bindPassword := util.Config.LdapBindPassword
	switch {
	case settings.LDAP.ClearBindPassword:
		bindPassword = ""
	case settings.LDAP.BindPassword != "":
		bindPassword = settings.LDAP.BindPassword
	}

	if err := persistAuthSettings(helpers.Store(r), settings, bindPassword); err != nil {
		helpers.WriteErrorStatus(w, "Can not save auth settings", http.StatusInternalServerError)
		return
	}

	applyAuthSettings(settings, bindPassword)

	helpers.WriteJSON(w, http.StatusOK, getCurrentAuthSettings())
}

func testLDAPSettings(w http.ResponseWriter, r *http.Request) {
	var request ldapTestRequest
	if !helpers.Bind(w, r, &request) {
		return
	}

	bindPassword := util.Config.LdapBindPassword
	switch {
	case request.LDAP.ClearBindPassword:
		bindPassword = ""
	case request.LDAP.BindPassword != "":
		bindPassword = request.LDAP.BindPassword
	}

	cfg := ldapRequestToRuntimeConfig(request.LDAP, bindPassword)
	cfg.Enabled = true

	if err := testLDAPReaderConnection(cfg); err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := ldapTestResponse{
		Message: "LDAP connection successful.",
	}

	if request.Login != "" || request.Password != "" {
		if request.Login == "" || request.Password == "" {
			helpers.WriteErrorStatus(w, "Both login and password are required for LDAP user validation", http.StatusBadRequest)
			return
		}

		user, err := findLDAPUserWithConfig(cfg, request.Login, request.Password)
		if err != nil {
			helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
			return
		}

		if user == nil {
			helpers.WriteErrorStatus(w, "LDAP search returned no entries", http.StatusNotFound)
			return
		}

		response.Message = "LDAP bind, search and user authentication succeeded."
		response.User = &ldapTestUserPreview{
			Name:     user.Name,
			Username: user.Username,
			Email:    user.Email,
		}
	}

	helpers.WriteJSON(w, http.StatusOK, response)
}
