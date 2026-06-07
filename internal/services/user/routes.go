package user

import (
	"encoding/json"
	"fmt"
	"megome/config"
	"megome/internal/services/auth"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	userStore    types.UserStore
	refreshStore types.RefreshTokenStore
}

func NewHandler(userStore types.UserStore, refreshStore types.RefreshTokenStore) *Handler {
	return &Handler{userStore: userStore, refreshStore: refreshStore}
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

func (h *Handler) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CALLBACK HIT")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("callback reached"))

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	fmt.Println("STATE:", state)
	fmt.Println("CODE:", code)

	if state != "random-state" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid state"))
		return
	}

	oauthConfig := auth.NewGoogleOAuthConfig()

	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	fmt.Println("ACCESS TOKEN:", token.AccessToken)
	fmt.Println("REFRESH TOKEN:", token.RefreshToken)
	fmt.Println("EXPIRY:", token.Expiry)

	client := oauthConfig.Client(r.Context(), token)

	resp, err := client.Get(
		"https://www.googleapis.com/oauth2/v2/userinfo",
	)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		Locale        string `json:"locale"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	fmt.Printf("GOOGLE USER: %+v\n", googleUser)

	utils.WriteJSON(w, http.StatusOK, googleUser)
}
