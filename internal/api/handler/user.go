package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"megome/internal/config"
	"megome/internal/domain/emailverification"
	"megome/internal/domain/passwordforgot"
	"megome/internal/domain/profile"
	"megome/internal/domain/refreshtoken"
	"megome/internal/domain/user"
	"megome/internal/pkg/auth"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/mailer"
	"megome/internal/pkg/validator"
	"net/http"
	"time"

	playvalidator "github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

type UserHandler struct {
	userStore         *user.Repository
	profileStore      *profile.Repository
	refreshStore      *refreshtoken.Repository
	emailService      *mailer.Service
	passwordForgot    *passwordforgot.Repository
	emailVerification *emailverification.Repository
}

func NewUserHandler(userStore *user.Repository, profileStore *profile.Repository, refreshStore *refreshtoken.Repository, emailService *mailer.Service, passwordForgot *passwordforgot.Repository, emailVerification *emailverification.Repository) *UserHandler {
	return &UserHandler{userStore: userStore, profileStore: profileStore, refreshStore: refreshStore, emailService: emailService, passwordForgot: passwordForgot, emailVerification: emailVerification}
}

func (h *UserHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/auth/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/auth/register", h.handleRegister).Methods("POST")
	router.HandleFunc("/auth/verify", h.handleVerify).Methods("GET")
	router.HandleFunc("/auth/logout", h.handleLogout).Methods("POST")
	router.HandleFunc("/auth/google", h.handleGoogleLogin).Methods("GET")
	router.HandleFunc("/auth/google/callback", h.handleGoogleCallback).Methods("GET")
	router.HandleFunc("/auth/forgot-pass", h.handleForgotPassword).Methods("POST")
	router.HandleFunc("/auth/change-forgot-pass", h.handleResetPassword).Methods("POST")
	router.HandleFunc("/auth/verify-email", h.handleVerifyEmail).Methods("POST")
	router.HandleFunc("/auth/resend-otp", h.handleResendOTP).Methods("POST")
}

func (h *UserHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload user.LoginUserPayload
	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	u, err := h.userStore.GetUserByEmailOrUsername(payload.EmailOrUsername)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid email or password"))
		return
	}

	if !auth.ComparePasswords(u.Password, []byte(payload.Password)) {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid email or password"))
		return
	}

	if u.EmailVerifiedAt == nil {
		httputil.WriteJSON(w, http.StatusForbidden, map[string]any{
			"error": "email not verified",
			"email": u.Email,
		})
		return
	}

	at, rt, err := h.getTokens(u.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := refreshtoken.AuthResponse{
		Success:      true,
		Message:      "Account was successfully logged in!",
		AccessToken:  at,
		RefreshToken: rt,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var payload user.RegisterUserPayload
	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}
	_, err := h.userStore.GetUserByEmail(payload.Email)
	if err == nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	hashedPassword, err := auth.HashedPassword(payload.Password)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	u, err := h.userStore.CreateUser(user.User{
		Username: payload.Username,
		Email:    payload.Email,
		Password: hashedPassword,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	otp, err := auth.GenerateOTP()
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	hash := sha256.Sum256([]byte(otp))
	otpHash := hex.EncodeToString(hash[:])

	err = h.emailVerification.SaveOTP(u.ID, u.Email, otpHash, time.Now().Add(emailverification.OTPExpiry))
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.emailService.SendVerifyEmail(u.Email, otp)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": "verification code sent to your email",
		"email":   u.Email,
	})
}

func (h *UserHandler) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var payload user.VerifyEmailPayload

	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	hash := sha256.Sum256([]byte(payload.OTP))
	otpHash := hex.EncodeToString(hash[:])

	userId, err := h.emailVerification.VerifyOTP(payload.Email, otpHash)
	if err != nil {
		if errors.Is(err, emailverification.ErrTooManyAttempts) {
			httputil.WriteError(w, http.StatusTooManyRequests, err)
			return
		}
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.emailVerification.MarkVerified(userId); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	at, rt, err := h.getTokens(userId)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := refreshtoken.AuthResponse{
		Success:      true,
		Message:      "Email verified successfully",
		AccessToken:  at,
		RefreshToken: rt,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) handleResendOTP(w http.ResponseWriter, r *http.Request) {
	var payload user.ResendOTPPayload

	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	u, err := h.userStore.GetUserByEmail(payload.Email)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, fmt.Errorf("user not found"))
		return
	}

	if u.EmailVerifiedAt != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("email already verified"))
		return
	}

	lastSent, err := h.emailVerification.LastOTPSentAt(u.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if remaining := emailverification.RemainingCooldown(lastSent, time.Now(), emailverification.ResendCooldown); remaining > 0 {
		httputil.WriteJSON(w, http.StatusConflict, map[string]any{
			"error":             "please wait before requesting another code",
			"retryAfterSeconds": remaining,
		})
		return
	}

	otp, err := auth.GenerateOTP()
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	hash := sha256.Sum256([]byte(otp))
	otpHash := hex.EncodeToString(hash[:])

	if err := h.emailVerification.DeleteOTPs(u.ID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.emailVerification.SaveOTP(u.ID, u.Email, otpHash, time.Now().Add(emailverification.OTPExpiry)); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.emailService.SendVerifyEmail(u.Email, otp); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "verification code sent to your email",
		"email":   u.Email,
	})
}

func (h *UserHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	refreshToken := httputil.GetTokenFromRequest(r)
	if refreshToken == "" {
		permissionDenied(w, "invalid token")
		return
	}

	err := h.refreshStore.LogoutUser(refreshToken)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := refreshtoken.AuthResponse{
		Success: true,
		Message: "User successfully logged out!",
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) handleVerify(w http.ResponseWriter, r *http.Request) {
	accesstoken, err := r.Cookie("Authentication")
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, fmt.Errorf("missing access cookie"))
		return
	}
	hasErr := auth.VerifyToken(accesstoken.Value)
	if hasErr != nil {
		httputil.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid access token"))
		return
	}

	resp := refreshtoken.AuthResponse{
		Success: true,
		Message: "access-token is valid",
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) getTokens(userId int) (string, string, error) {
	secret := []byte(config.Envs.JWTSecret)
	accessToken, err := auth.CreateJWT(secret, userId)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := h.refreshStore.CreateRefreshToken(userId)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (h *UserHandler) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := httputil.GenerateRandomToken("oauth_")
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		MaxAge:   600,
		SameSite: http.SameSiteLaxMode,
	})

	oauthConfig := auth.NewGoogleOAuthConfig()
	url := oauthConfig.AuthCodeURL(state)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func permissionDenied(w http.ResponseWriter, m string) {
	httputil.WriteError(w, http.StatusUnauthorized, fmt.Errorf("permission denied %v", m))
}

func (h *UserHandler) handleGoogleCallback(
	w http.ResponseWriter,
	r *http.Request,
) {
	queryState := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if code == "" {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			fmt.Errorf("missing authorization code"),
		)
		return
	}

	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value == "" || cookie.Value != queryState {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			fmt.Errorf("invalid state parameter"),
		)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "oauth_state", Value: "", MaxAge: -1, Path: "/",
	})

	oauthConfig := auth.NewGoogleOAuthConfig()

	token, err := oauthConfig.Exchange(
		r.Context(),
		code,
	)

	if err != nil {
		httputil.WriteError(
			w,
			http.StatusBadRequest,
			err,
		)
		return
	}

	googleUser, err := getGoogleUser(
		r.Context(),
		oauthConfig,
		token,
	)

	if err != nil {
		httputil.WriteError(
			w,
			http.StatusInternalServerError,
			err,
		)
		return
	}

	if !googleUser.VerifiedEmail {
		httputil.WriteError(
			w,
			http.StatusUnauthorized,
			fmt.Errorf("email not verified"),
		)
		return
	}

	account, err := h.userStore.GetOAuthAccount(
		"google",
		googleUser.ID,
	)

	var u *user.User

	if err == nil {
		u, err = h.userStore.GetUserByID(
			account.UserID,
		)

		if err != nil {
			httputil.WriteError(
				w,
				http.StatusInternalServerError,
				err,
			)
			return
		}

	} else {
		u, err = h.userStore.GetUserByEmail(
			googleUser.Email,
		)

		if err != nil {
			u, err = h.userStore.CreateUser(
				user.User{
					Username: googleUser.Email,
					Email:    googleUser.Email,
					Password: "",
				},
			)

			if err != nil {
				httputil.WriteError(w, http.StatusInternalServerError, err)
				return
			}
			err = h.userStore.MarkEmailVerified(u.ID)
			if err != nil {
				httputil.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			err = h.profileStore.UpsertOAuthProfile(
				profile.Profile{
					UserID:       u.ID,
					FirstName:    googleUser.GivenName,
					LastName:     googleUser.FamilyName,
					ProfileImage: googleUser.Picture,
				},
			)

			if err != nil {
				httputil.WriteError(
					w,
					http.StatusInternalServerError,
					err,
				)
				return
			}
		}

		email := googleUser.Email

		err = h.userStore.CreateOAuthAccount(
			user.OAuthAccount{
				UserID:         u.ID,
				Provider:       "google",
				ProviderUserID: googleUser.ID,
				Email:          &email,
			},
		)

		if err != nil {
			httputil.WriteError(
				w,
				http.StatusInternalServerError,
				err,
			)
			return
		}
	}

	// Mark email as verified for all Google OAuth users (idempotent).
	// Covers new users and existing email/password users linking Google.
	if err := h.userStore.MarkEmailVerified(u.ID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	accessToken, refreshToken, err := h.getTokens(
		u.ID,
	)

	if err != nil {
		httputil.WriteError(
			w,
			http.StatusInternalServerError,
			err,
		)
		return
	}
	redirectURL := fmt.Sprintf(
		"%s/auth/google/success?access_token=%s&refresh_token=%s",
		config.Envs.FrontendUrl,
		accessToken,
		refreshToken,
	)

	http.Redirect(
		w,
		r,
		redirectURL,
		http.StatusTemporaryRedirect,
	)
}

func getGoogleUser(ctx context.Context, oauthConfig *oauth2.Config, token *oauth2.Token,
) (*user.GoogleUser, error) {
	client := oauthConfig.Client(ctx, token)

	resp, err := client.Get(
		"https://www.googleapis.com/oauth2/v2/userinfo",
	)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"google returned status %d",
			resp.StatusCode,
		)
	}

	var u user.GoogleUser

	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}

	return &u, nil
}

func (h *UserHandler) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var payload user.ForgotPassChangePayload

	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	err := h.passwordForgot.ChangePassword(payload.Token, payload.Password)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := refreshtoken.AuthResponse{
		Success: true,
		Message: "password successfully reset",
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var payload user.ForgotPassPayload

	if err := httputil.ParseJSON(r, &payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := validator.Validate.Struct(payload); err != nil {
		errors := err.(playvalidator.ValidationErrors)
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	u, err := h.userStore.GetUserByEmail(payload.Email)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, fmt.Errorf("user not found"))
		return
	}

	resetToken := uuid.NewString()

	err = h.passwordForgot.SavePasswordResetToken(
		u.ID,
		resetToken,
		time.Now().Add(15*time.Minute),
	)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resetURL := fmt.Sprintf(
		"%s/auth/reset-password?token=%s",
		config.Envs.FrontendUrl,
		resetToken,
	)
	fmt.Println("DEBUG USER EMAIL", u.Email)
	err = h.emailService.SendResetPassword(u.Email, resetURL)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "password reset email sent",
	})
}
