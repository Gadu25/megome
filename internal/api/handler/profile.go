package handler

import (
	"fmt"
	"io"
	"megome/internal/domain/profile"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/storage"
	"megome/internal/pkg/validator"
	"net/http"

	"github.com/gorilla/mux"
)

type ProfileHandler struct {
	profileStore *profile.Repository
	userStore    *user.Repository
	r2Client     *storage.R2Client
}

type ProfileResponse struct {
	Message string          `json:"message"`
	Profile *profile.Profile `json:"profile"`
}

func NewProfileHandler(profileStore *profile.Repository, userStore *user.Repository, r2Client *storage.R2Client) *ProfileHandler {
	return &ProfileHandler{profileStore: profileStore, userStore: userStore, r2Client: r2Client}
}

func (h *ProfileHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/profile", middleware.WithJWTAuth(h.handleViewProfiles, h.userStore)).Methods("GET")
	router.HandleFunc("/profile", middleware.WithJWTAuth(h.handleUpdateProfile, h.userStore)).Methods("POST")
}

func (h *ProfileHandler) handleViewProfiles(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	p, err := h.profileStore.GetProfile(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := ProfileResponse{
		Message: "Profile fetched successfully",
		Profile: p,
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *ProfileHandler) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())

	payload := profile.MakeProfilePayload{
		Bio:       r.FormValue("bio"),
		FirstName: r.FormValue("firstName"),
		LastName:  r.FormValue("lastName"),
		Tagline:   r.FormValue("tagline"),
		Title:     r.FormValue("title"),
		Birthday:  httputil.PointerFromString(r.FormValue("birthday")),
		Phone:     r.FormValue("phone"),
		Website:   r.FormValue("website"),
		Location:  r.FormValue("location"),
	}
	var profileImageKey string

	file, handler, _ := r.FormFile("profileImage")
	if file != nil {
		if handler.Size > 1<<20 {
			httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("file too large (max 1MB)"))
			return
		}

		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("failed to read file: %w", err))
			return
		}

		key, err := h.r2Client.UploadImage(r.Context(), data,
			fmt.Sprintf("profiles/%d", userID),
			"avatar",
		)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to upload image: %w", err))
			return
		}

		existing, err := h.profileStore.GetProfile(userID)
		if err == nil && existing != nil && existing.ProfileImage != "" {
			_ = h.r2Client.DeleteObject(r.Context(), existing.ProfileImage)
		}

		profileImageKey = key
	}

	if err := validator.Validate.Struct(payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", err))
		return
	}

	err := h.profileStore.MakeProfile(profile.Profile{
		UserID:       userID,
		Tagline:      payload.Tagline,
		Bio:          payload.Bio,
		FirstName:    payload.FirstName,
		LastName:     payload.LastName,
		Title:        payload.Title,
		Birthday:     payload.Birthday,
		Phone:        payload.Phone,
		Website:      payload.Website,
		Location:     payload.Location,
		ProfileImage: profileImageKey,
	})

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	p, err := h.profileStore.GetProfile(userID)

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := ProfileResponse{
		Message: "Profile updated successfully",
		Profile: p,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
