package api

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image/png"
	"net/http"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	proApi "github.com/semaphoreui/semaphore/pro/api"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"

	"github.com/pquerna/otp/totp"
)

const (
	totpSetupCookieName     = "semaphore_totp_setup"
	totpSetupFallbackWindow = 24 * time.Hour
)

func getSession(r *http.Request) (*db.Session, bool) {
	// fetch session from cookie
	cookie, err := r.Cookie("semaphore")
	if err != nil {
		return nil, false
	}

	value := make(map[string]any)
	if err = util.Cookie.Decode("semaphore", cookie.Value, &value); err != nil {
		//w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}

	user, ok := value["user"]
	sessionVal, okSession := value["session"]
	if !ok || !okSession {
		//w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}

	userID := user.(int)
	sessionID := sessionVal.(int)

	// fetch session
	session, err := helpers.Store(r).GetSession(userID, sessionID)

	if err != nil {
		//w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}

	if time.Since(session.LastActive).Hours() > 7*24 {
		// more than week old unused session
		// destroy.
		if err = helpers.Store(r).ExpireSession(userID, sessionID); err != nil {
			// it is internal error, it doesn't concern the user
			log.Error(err)
		}

		return nil, false
	}

	return &session, true

}

type totpRequestBody struct {
	Passcode string `json:"passcode"`
}

type totpRecoveryRequestBody struct {
	RecoveryCode string `json:"recovery_code"`
}

type totpSetupCookiePayload struct {
	UserID       int    `json:"user"`
	SessionID    int    `json:"session"`
	RecoveryCode string `json:"recovery_code,omitempty"`
}

type totpSetupResponse struct {
	Secret       string `json:"secret"`
	Issuer       string `json:"issuer"`
	AccountName  string `json:"account_name"`
	RecoveryCode string `json:"recovery_code,omitempty"`
	QRCode       string `json:"qr_code,omitempty"`
}

func clearTotpSetupCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     totpSetupCookieName,
		Value:    "",
		Expires:  tz.Now().Add(-7 * 24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	})
}

func writeTotpSetupCookie(w http.ResponseWriter, session db.Session, recoveryCode string) error {
	encoded, err := util.Cookie.Encode(totpSetupCookieName, totpSetupCookiePayload{
		UserID:       session.UserID,
		SessionID:    session.ID,
		RecoveryCode: recoveryCode,
	})
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     totpSetupCookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
	})

	return nil
}

func readTotpSetupCookie(r *http.Request) (*totpSetupCookiePayload, bool) {
	cookie, err := r.Cookie(totpSetupCookieName)
	if err != nil {
		return nil, false
	}

	var payload totpSetupCookiePayload
	if err = util.Cookie.Decode(totpSetupCookieName, cookie.Value, &payload); err != nil {
		return nil, false
	}

	return &payload, true
}

func ensureUserTotpEnrollment(store db.Store, user db.User) (db.User, string, bool, error) {
	if !util.GetTotpConfig().Enabled || user.Totp != nil {
		return user, "", false, nil
	}

	issuer := "Semaphore"
	if util.GetTotpConfig().Issuer != "" {
		issuer = util.GetTotpConfig().Issuer
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: user.Email,
	})
	if err != nil {
		return user, "", false, err
	}

	var code string
	var hash string

	if util.GetTotpConfig().AllowRecovery {
		code, hash, err = util.GenerateRecoveryCode()
		if err != nil {
			return user, "", false, err
		}
	}

	newTotp, err := store.AddTotpVerification(user.ID, key.URL(), hash)
	if err != nil {
		return user, "", false, err
	}

	newTotp.RecoveryCode = code
	user.Totp = &newTotp

	return user, code, true, nil
}

func getTotpSetupState(r *http.Request) (*db.Session, *db.User, *totpSetupCookiePayload, bool) {
	session, ok := getSession(r)
	if !ok || session.VerificationMethod != db.SessionVerificationTotp || session.Verified {
		return nil, nil, nil, false
	}

	user, err := helpers.Store(r).GetUser(session.UserID)
	if err != nil || user.Totp == nil {
		return nil, nil, nil, false
	}

	payload, ok := readTotpSetupCookie(r)
	if ok && payload.UserID == session.UserID && payload.SessionID == session.ID {
		return session, &user, payload, true
	}

	// Grace window for newly provisioned TOTP secrets. This also recovers users
	// who first logged in while the setup cookie was not yet being written.
	if tz.Now().Sub(user.Totp.Created) <= totpSetupFallbackWindow {
		return session, &user, &totpSetupCookiePayload{
			UserID:    session.UserID,
			SessionID: session.ID,
		}, true
	}

	return nil, nil, nil, false
}

func getTotpSetup(w http.ResponseWriter, r *http.Request) {
	_, user, payload, ok := getTotpSetupState(r)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	setup, err := buildTotpSetupResponse(*user, payload.RecoveryCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, setup)
}

func getTotpSetupQr(w http.ResponseWriter, r *http.Request) {
	_, user, _, ok := getTotpSetupState(r)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	key, err := otp.NewKeyFromURL(user.Totp.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	image, err := key.Image(256, 256)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, image)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "image/png")
	_, err = w.Write(buf.Bytes())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func buildTotpSetupResponse(user db.User, recoveryCode string) (totpSetupResponse, error) {
	key, err := otp.NewKeyFromURL(user.Totp.URL)
	if err != nil {
		return totpSetupResponse{}, err
	}

	image, err := key.Image(256, 256)
	if err != nil {
		return totpSetupResponse{}, err
	}

	var buf bytes.Buffer
	if err = png.Encode(&buf, image); err != nil {
		return totpSetupResponse{}, err
	}

	return totpSetupResponse{
		Secret:       key.Secret(),
		Issuer:       key.Issuer(),
		AccountName:  key.AccountName(),
		RecoveryCode: recoveryCode,
		QRCode:       "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

// recoverySession handles the recovery of a user session using a recovery code.
// It validates the recovery code provided by the user and, if valid, verifies the session.
// If the recovery code is invalid or recovery is not allowed, it returns an appropriate HTTP status code.
//
// HTTP Request:
// - Method: POST
// - Body: JSON object containing the recovery code (e.g., {"recovery_code": "code"}).
//
// Responses:
// - 204 No Content: Recovery successful, session verified.
// - 400 Bad Request: Invalid request body or user does not have TOTP enabled.
// - 401 Unauthorized: Invalid recovery code or session not found.
// - 403 Forbidden: TOTP recovery is disabled.
// - 500 Internal Server Error: An unexpected error occurred.
//
// Preconditions:
// - The session must exist and be valid.
// - TOTP recovery must be enabled in the configuration.
//
// Parameters:
// - w: The HTTP response writer.
// - r: The HTTP request.
func recoverySession(w http.ResponseWriter, r *http.Request) {
	session, ok := getSession(r)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch session.VerificationMethod {
	case db.SessionVerificationTotp:
		if !util.GetTotpConfig().Enabled || !util.GetTotpConfig().AllowRecovery {
			helpers.WriteErrorStatus(w, "TOTP_DISABLED", http.StatusForbidden)
			return
		}

		var body totpRecoveryRequestBody
		if !helpers.Bind(w, r, &body) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		store := helpers.Store(r)

		user, err := store.GetUser(session.UserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user.Totp == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !util.VerifyRecoveryCode(body.RecoveryCode, user.Totp.RecoveryHash) {
			helpers.WriteErrorStatus(w, "INVALID_RECOVERY_CODE", http.StatusUnauthorized)
			return
		}

		err = store.DeleteTotpVerification(user.ID, user.Totp.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = store.VerifySession(session.UserID, session.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		clearTotpSetupCookie(w)
		w.WriteHeader(http.StatusNoContent)
	case db.SessionVerificationNone:
		clearTotpSetupCookie(w)
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func verifySession(w http.ResponseWriter, r *http.Request) {
	session, ok := getSession(r)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch session.VerificationMethod {
	case db.SessionVerificationEmail:
		proApi.VerifySessionByEmail(session, w, r)
		return

	case db.SessionVerificationTotp:
		if !util.GetTotpConfig().Enabled {
			helpers.WriteErrorStatus(w, "TOTP_DISABLED", http.StatusForbidden)
			return
		}

		var body totpRequestBody
		if !helpers.Bind(w, r, &body) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		user, err := helpers.Store(r).GetUser(session.UserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user.Totp == nil {
			helpers.WriteErrorStatus(w, "TOTP_SETUP_REQUIRED", http.StatusUnauthorized)
			return
		}

		key, err := otp.NewKeyFromURL(user.Totp.URL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !totp.Validate(body.Passcode, key.Secret()) {
			helpers.WriteErrorStatus(w, "INVALID_PASSCODE", http.StatusUnauthorized)
			return
		}

		err = helpers.Store(r).VerifySession(session.UserID, session.ID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		clearTotpSetupCookie(w)

	case db.SessionVerificationNone:
		clearTotpSetupCookie(w)
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func authenticationHandler(w http.ResponseWriter, r *http.Request) (ok bool, req *http.Request) {
	var userID int

	req = r

	authHeader := strings.ToLower(r.Header.Get("authorization"))

	if len(authHeader) > 0 && strings.Contains(authHeader, "bearer") {
		token, err := helpers.Store(r).GetAPIToken(strings.Replace(authHeader, "bearer ", "", 1))

		if err != nil {
			if !errors.Is(err, db.ErrNotFound) {
				log.Error(err)
			}

			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID = token.UserID
	} else {
		session, found := getSession(r)

		if !found {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !session.IsVerified() {
			switch session.VerificationMethod {
			case db.SessionVerificationEmail:
				helpers.WriteErrorStatus(w, "EMAIL_OTP_REQUIRED", http.StatusUnauthorized)
			case db.SessionVerificationTotp:
				helpers.WriteErrorStatus(w, "TOTP_REQUIRED", http.StatusUnauthorized)
			default:
				helpers.WriteErrorStatus(w, "SESSION_NOT_VERIFIED", http.StatusUnauthorized)
			}
			return
		}

		userID = session.UserID

		if err := helpers.Store(r).TouchSession(userID, session.ID); err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	user, err := helpers.Store(r).GetUser(userID)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			// internal error
			log.Error(err)
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ok = true
	req = helpers.SetContextValue(r, "user", &user)
	return
}

// nolint: gocyclo
func authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok, r := authenticationHandler(w, r)
		if ok {
			next.ServeHTTP(w, r)
		}
	})
}

// nolint: gocyclo
func authenticationWithStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store := helpers.Store(r)

		var ok bool

		db.StoreSession(store, r.URL.String(), func() {
			ok, r = authenticationHandler(w, r)
		})

		if ok {
			next.ServeHTTP(w, r)
		}
	})
}

func adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := helpers.GetFromContext(r, "user").(*db.User)

		if !user.Admin {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
