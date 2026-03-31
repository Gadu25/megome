package user

import (
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

	u, err := h.userStore.GetUserByEmail(payload.Email)
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

	utils.SetRefreshTokenCookie(w, rt)
	utils.SetAccessTokenCookie(w, at)
	resp := types.AuthResponse{
		Success: true,
		Message: "Account was successfully logged in!",
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
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  hashedPassword,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	at, rt, err := h.getTokens(user.ID)

	utils.SetRefreshTokenCookie(w, rt)
	utils.SetAccessTokenCookie(w, at)

	resp := types.AuthResponse{
		Success: true,
		Message: "Your account is successfully registered!",
	}
	utils.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("Error getting cookie %v", err))
		return
	}

	err = h.refreshStore.LogoutUser(cookie.Value)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := types.AuthResponse{
		Success: true,
		Message: "User successfully logged out!",
	}
	utils.ClearRefreshTokenCookie(w)
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
