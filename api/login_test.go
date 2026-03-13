package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	boltstore "github.com/semaphoreui/semaphore/db/bolt"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseClaim(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	res, ok := parseClaim("email | {{ .id }}@test.com", claims)

	assert.True(t, ok, "parseClaim should succeed")
	assert.Equal(t, "1234567@test.com", res, "Result should be formatted correctly")
}

func TestParseClaim2(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	res, ok := parseClaim("username", claims)

	assert.True(t, ok, "parseClaim should succeed")
	assert.Equal(t, claims["username"], res, "Result should match username claim")
}

func TestParseClaim3(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	_, ok := parseClaim("email", claims)

	assert.False(t, ok, "parseClaim should fail for empty email")
}

func TestParseClaim4(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       1234567,
	}

	_, ok := parseClaim("|", claims)

	assert.False(t, ok, "parseClaim should fail for invalid pattern")
}

func TestParseClaim5(t *testing.T) {
	claims := map[string]any{
		"username": "fiftin",
		"email":    "",
		"id":       123456757343.0,
	}

	prepareClaims(claims)

	res, ok := parseClaim("{{ .id }}", claims)

	assert.True(t, ok, "parseClaim should succeed")
	assert.Equal(t, "123456757343", res, "Result should match formatted ID")
}

func TestGenerateStateOauthCookie(t *testing.T) {
	w := httptest.NewRecorder()
	returnPath := "/dashboard"

	stateStr := generateStateOauthCookie(w, returnPath)

	// Test 1: Verify returned state is valid base64
	stateBytes, err := base64.URLEncoding.DecodeString(stateStr)
	assert.NoError(t, err, "Returned state should be valid base64")

	// Test 2: Verify state contains valid JSON
	var state oAuthState
	err = json.Unmarshal(stateBytes, &state)
	assert.NoError(t, err, "State should contain valid JSON")

	// Test 3: Verify return path is preserved
	assert.Equal(t, returnPath, state.Return, "Return path should be preserved")

	// Test 4: Verify CSRF token is not empty
	assert.NotEmpty(t, state.Csrf, "CSRF token should not be empty")

	// Test 5: Verify CSRF token is valid base64
	_, err = base64.URLEncoding.DecodeString(state.Csrf)
	assert.NoError(t, err, "CSRF token should be valid base64")

	// Test 6: Verify cookie is set
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies, "At least one cookie should be set")

	// Test 7: Verify cookie has correct name
	var oauthCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "oauthstate" {
			oauthCookie = cookie
			break
		}
	}
	assert.NotNil(t, oauthCookie, "Cookie 'oauthstate' should be set")

	// Test 8: Verify cookie value matches CSRF token in state
	assert.Equal(t, state.Csrf, oauthCookie.Value, "Cookie value should match CSRF token")

	// Test 9: Verify cookie has expiration set (should be ~365 days)
	assert.False(t, oauthCookie.Expires.IsZero(), "Cookie expiration should be set")

	expectedExpiration := time.Now().Add(365 * 24 * time.Hour)
	timeDiff := oauthCookie.Expires.Sub(expectedExpiration)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	// Allow 5 seconds tolerance for test execution time
	assert.LessOrEqual(t, timeDiff, 5*time.Second, "Cookie expiration should be within 5 seconds of expected")
}

func TestGenerateStateOauthCookieEmptyReturnPath(t *testing.T) {
	w := httptest.NewRecorder()
	returnPath := ""

	stateStr := generateStateOauthCookie(w, returnPath)

	// Decode and verify state
	stateBytes, err := base64.URLEncoding.DecodeString(stateStr)
	assert.NoError(t, err, "Returned state should be valid base64")

	var state oAuthState
	err = json.Unmarshal(stateBytes, &state)
	assert.NoError(t, err, "State should contain valid JSON")

	// Verify empty return path is preserved
	assert.Empty(t, state.Return, "Return path should be empty")
}

func TestGenerateStateOauthCookieUniqueness(t *testing.T) {
	// Generate two states and verify they have different CSRF tokens
	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()

	state1Str := generateStateOauthCookie(w1, "/path1")
	state2Str := generateStateOauthCookie(w2, "/path2")

	// Decode states
	state1Bytes, err1 := base64.URLEncoding.DecodeString(state1Str)
	state2Bytes, err2 := base64.URLEncoding.DecodeString(state2Str)
	assert.NoError(t, err1, "First state should be valid base64")
	assert.NoError(t, err2, "Second state should be valid base64")

	var state1, state2 oAuthState
	err1 = json.Unmarshal(state1Bytes, &state1)
	err2 = json.Unmarshal(state2Bytes, &state2)
	assert.NoError(t, err1, "First state should be valid JSON")
	assert.NoError(t, err2, "Second state should be valid JSON")

	// Verify CSRF tokens are different
	assert.NotEqual(t, state1.Csrf, state2.Csrf, "Multiple calls should generate different CSRF tokens")

	// Verify states are different
	assert.NotEqual(t, state1Str, state2Str, "Multiple calls should generate different state strings")
}

func TestGetLDAPFallbackUsername(t *testing.T) {
	assert.Equal(t, "richard.silva", getLDAPFallbackUsername(`CLICKIP\\richard.silva`, ""))
	assert.Equal(t, "richard.silva", getLDAPFallbackUsername("richard.silva@clickip.com.br", ""))
	assert.Equal(t, "richard.silva", getLDAPFallbackUsername("", "richard.silva@clickip.com.br"))
}

func TestParseLDAPClaimsUsesStableFallbackUsername(t *testing.T) {
	claims := map[string]any{
		"mail": "richard.silva@clickip.com.br",
	}

	res, err := parseLDAPClaims(claims, &util.LdapMappings{
		UID:  "uid",
		Mail: "mail",
		CN:   "displayName",
	}, "richard.silva")

	assert.NoError(t, err)
	assert.Equal(t, "richard.silva", res.username)
	assert.Equal(t, "richard.silva@clickip.com.br", res.email)
}

func TestNormalizeLDAPConfigUsesActiveDirectoryMappings(t *testing.T) {
	cfg := normalizeLDAPConfig(ldapRuntimeConfig{
		SearchFilter: defaultADSearchFilter,
		Mappings: util.LdapMappings{
			DN:   "dn",
			UID:  "uid",
			CN:   "cn",
			Mail: "mail",
		},
	})

	assert.Equal(t, defaultADUIDMapping, cfg.Mappings.UID)
	assert.Equal(t, defaultADCNMapping, cfg.Mappings.CN)
	assert.Equal(t, defaultADDNMapping, cfg.Mappings.DN)
	assert.Equal(t, defaultADMailMapping, cfg.Mappings.Mail)
}

func TestGetLDAPSearchCandidates(t *testing.T) {
	assert.Equal(t, []string{"richard.silva"}, getLDAPSearchCandidates("richard.silva"))
	assert.Equal(t, []string{"richard.silva@clickip.local", "richard.silva"}, getLDAPSearchCandidates("richard.silva@clickip.local"))
	assert.Equal(t, []string{`CLICKIP\richard.silva`, "richard.silva"}, getLDAPSearchCandidates(`CLICKIP\richard.silva`))
}

func TestGetLDAPBindCandidates(t *testing.T) {
	cfg := ldapRuntimeConfig{
		SearchDN: "DC=clickip,DC=local",
	}

	assert.Equal(t, []string{
		"richard.silva",
		"richard.silva@clickip.local",
		`CLICKIP\richard.silva`,
	}, getLDAPBindCandidates(cfg, "richard.silva"))

	assert.Equal(t, []string{
		"richard.silva@clickip.local",
		"richard.silva",
	}, getLDAPBindCandidates(cfg, "richard.silva@clickip.local"))
}

func TestLoginByLDAPCreatesExternalUserWhenMissing(t *testing.T) {
	store := boltstore.CreateTestStore()

	ldapUser := db.User{
		Username: "richard.silva",
		Name:     "Richard Barbosa Silva",
		Email:    "richard.silva@clickip.com.br",
		External: true,
	}

	user, err := loginByLDAP(store, ldapUser)
	require.NoError(t, err)

	assert.True(t, user.External)
	assert.Equal(t, ldapUser.Username, user.Username)
	assert.Equal(t, ldapUser.Email, user.Email)

	found, err := store.GetUserByLoginOrEmail(ldapUser.Username, ldapUser.Email)
	require.NoError(t, err)
	assert.True(t, found.External)
	assert.Equal(t, ldapUser.Name, found.Name)
}

func TestLoginByLDAPLinksExistingLocalUser(t *testing.T) {
	store := boltstore.CreateTestStore()

	localUser, err := store.CreateUser(db.UserWithPwd{
		Pwd: "123456",
		User: db.User{
			Username: "richard.silva",
			Name:     "Richard Local",
			Email:    "old.richard@clickip.com.br",
			Admin:    true,
			Alert:    true,
			Pro:      true,
		},
	})
	require.NoError(t, err)

	ldapUser := db.User{
		Username: "richard.silva",
		Name:     "Richard Barbosa Silva",
		Email:    "richard.silva@clickip.com.br",
		External: true,
	}

	user, err := loginByLDAP(store, ldapUser)
	require.NoError(t, err)

	assert.Equal(t, localUser.ID, user.ID)
	assert.True(t, user.External)
	assert.Equal(t, ldapUser.Name, user.Name)
	assert.Equal(t, ldapUser.Email, user.Email)
	assert.True(t, user.Admin)
	assert.True(t, user.Alert)
	assert.True(t, user.Pro)

	found, err := store.GetUser(localUser.ID)
	require.NoError(t, err)
	assert.True(t, found.External)
	assert.Equal(t, ldapUser.Email, found.Email)
	assert.Equal(t, ldapUser.Name, found.Name)
	assert.True(t, found.Admin)
}
