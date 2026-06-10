package user

import (
	"context"
	"encoding/json"
	"fmt"
	"megome/config"
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

type Handler struct {
	userStore    types.UserStore
	profileStore types.ProfileStore
	refreshStore types.RefreshTokenStore
}

func NewHandler(userStore types.UserStore, profileStore types.ProfileStore, refreshStore types.RefreshTokenStore) *Handler {
	return &Handler{userStore: userStore, profileStore: profileStore, refreshStore: refreshStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/auth/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/auth/register", h.handleRegister).Methods("POST")
	router.HandleFunc("/auth/verify", h.handleVerify).Methods("GET")
	router.HandleFunc("/auth/logout", h.handleLogout).Methods("POST")
	router.HandleFunc("/auth/google", h.handleGoogleLogin).Methods("GET")
	router.HandleFunc("/auth/google/callback", h.handleGoogleCallback).Methods("GET")
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	var payload types.LoginUserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	u, err := h.userStore.GetUserByEmailOrUsername(payload.EmailOrUsername)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("Not found, invalid email or password"))
		return
	}

	if !auth.ComparePasswords(u.Password, []byte(payload.Password)) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("Not found, invalid email or password"))
		return
	}

	at, rt, err := h.getTokens(u.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	resp := types.AuthResponse{
		Success:      true,
		Message:      "Account was successfully logged in!",
		AccessToken:  at,
		RefreshToken: rt,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// get JSON payload
	var payload types.RegisterUserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	// validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}
	// check if the user exists
	_, err := h.userStore.GetUserByEmail(payload.Email)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	// if it doesn't we create user
	hashedPassword, err := auth.HashedPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	user, err := h.userStore.CreateUser(types.User{
		Username: payload.Username,
		Email:    payload.Email,
		Password: hashedPassword,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	at, rt, err := h.getTokens(user.ID)

	// utils.SetRefreshTokenCookie(w, rt)
	// utils.SetAccessTokenCookie(w, at)

	resp := types.AuthResponse{
		Success:      true,
		Message:      "Your account is successfully registered!",
		AccessToken:  at,
		RefreshToken: rt,
	}
	utils.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	refreshToken := utils.GetTokenFromRequest(r)
	if refreshToken == "" {
		permissionDenied(w, "invalid token")
		return
	}

	err := h.refreshStore.LogoutUser(refreshToken)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := types.AuthResponse{
		Success: true,
		Message: "User successfully logged out!",
	}
	// utils.ClearAllTokens(w)
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleVerify(w http.ResponseWriter, r *http.Request) {
	accesstoken, err := r.Cookie("Authentication")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("Error getting access cookie %v", err))
		return
	}
	hasErr := auth.VerifyToken(accesstoken.Value)
	if hasErr != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("Access token is invalid %v", err))
		return
	}
	fmt.Println("verified!")

	resp := types.AuthResponse{
		Success: true,
		Message: "access-token is valid",
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) getTokens(userId int) (string, string, error) {
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

func (h *Handler) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	// IMPORTANT: in production, generate random state and store in cookie/session
	state := "random-state"

	oauthConfig := auth.NewGoogleOAuthConfig()
	url := oauthConfig.AuthCodeURL(state)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func permissionDenied(w http.ResponseWriter, m string) {
	utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("permission denied %v", m))
}

func (h *Handler) handleGoogleCallback(
	w http.ResponseWriter,
	r *http.Request,
) {

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		utils.WriteError(
			w,
			http.StatusBadRequest,
			fmt.Errorf("missing oauth parameters"),
		)
		return
	}

	if state != "random-state" {
		utils.WriteError(
			w,
			http.StatusBadRequest,
			fmt.Errorf("invalid state"),
		)
		return
	}

	oauthConfig := auth.NewGoogleOAuthConfig()

	token, err := oauthConfig.Exchange(
		r.Context(),
		code,
	)

	if err != nil {
		utils.WriteError(
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
		utils.WriteError(
			w,
			http.StatusInternalServerError,
			err,
		)
		return
	}

	if !googleUser.VerifiedEmail {
		utils.WriteError(
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

	var user *types.User

	if err == nil {

		user, err = h.userStore.GetUserByID(
			account.UserID,
		)

		if err != nil {
			utils.WriteError(
				w,
				http.StatusInternalServerError,
				err,
			)
			return
		}

	} else {

		user, err = h.userStore.GetUserByEmail(
			googleUser.Email,
		)

		if err != nil {

			user, err = h.userStore.CreateUser(
				types.User{
					Username: googleUser.Email,
					Email:    googleUser.Email,
					Password: "",
				},
			)

			err = h.profileStore.UpsertOAuthProfile(
				types.Profile{
					UserID:       user.ID,
					FirstName:    googleUser.GivenName,
					LastName:     googleUser.FamilyName,
					ProfileImage: googleUser.Picture,
				},
			)

			if err != nil {
				utils.WriteError(
					w,
					http.StatusInternalServerError,
					err,
				)
				return
			}

			if err != nil {
				utils.WriteError(
					w,
					http.StatusInternalServerError,
					err,
				)
				return
			}
		}

		email := googleUser.Email

		err = h.userStore.CreateOAuthAccount(
			types.OAuthAccount{
				UserID:         user.ID,
				Provider:       "google",
				ProviderUserID: googleUser.ID,
				Email:          &email,
			},
		)

		if err != nil {
			utils.WriteError(
				w,
				http.StatusInternalServerError,
				err,
			)
			return
		}
	}

	accessToken, refreshToken, err := h.getTokens(
		user.ID,
	)

	if err != nil {
		utils.WriteError(
			w,
			http.StatusInternalServerError,
			err,
		)
		return
	}
	fmt.Println("LOG FRONTEND URL: ", config.Envs.FrontendUrl)
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
) (*types.GoogleUser, error) {

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

	var user types.GoogleUser

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
