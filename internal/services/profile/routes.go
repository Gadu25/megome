package profile

import (
	"fmt"
	"megome/internal/services/auth"
	"megome/internal/services/storage"
	"megome/internal/services/types"
	"megome/internal/services/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	profileStore types.ProfileStore
	userStore    types.UserStore
	r2Client     *storage.R2Client
}
type ProfileResponse struct {
	Message string         `json:"message"`
	Data    *types.Profile `json:"data"`
}

func NewHandler(profileStore types.ProfileStore, userStore types.UserStore, r2Client *storage.R2Client) *Handler {
	return &Handler{profileStore: profileStore, userStore: userStore, r2Client: r2Client}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/profile", auth.WithJWTAuth(h.handleViewProfiles, h.userStore)).Methods("GET")
	router.HandleFunc("/profile", auth.WithJWTAuth(h.handleUpdateProfile, h.userStore)).Methods("POST")
}

func (h *Handler) handleViewProfiles(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	profile, err := h.profileStore.GetProfile(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	resp := ProfileResponse{
		Message: "Profile fetched successfully",
		Data:    profile,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())

	// Parse multipart form (10MB max)
	// TODO: This should be updated to limit file sizes to less 1mb
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("failed to parse form: %w", err))
		return
	}

	payload := types.MakeProfilePayload{
		Bio:       r.FormValue("bio"),
		FirstName: r.FormValue("firstName"),
		LastName:  r.FormValue("lastName"),
		Phone:     r.FormValue("phone"),
		Website:   r.FormValue("website"),
		Location:  r.FormValue("location"),
	}

	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %w", err))
		return
	}

	var profileImageKey string
	file, header, err := r.FormFile("profileImage")
	if err == nil && file != nil {
		defer file.Close()

		key, err := storage.GenerateKey(fmt.Sprintf("profiles/%d", userID), header.Filename)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid file name: %w", err))
			return
		}

		err = h.r2Client.UploadFromReader(r.Context(), key, file, header.Size, header.Header.Get("Content-Type"))
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError,
				fmt.Errorf("failed to upload image: %w", err))
			return
		}

		profileImageKey = key
	} else {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Save profile to DB
	err = h.profileStore.MakeProfile(types.Profile{
		UserID:       userID,
		Bio:          payload.Bio,
		FirstName:    payload.FirstName,
		LastName:     payload.LastName,
		Phone:        payload.Phone,
		Website:      payload.Website,
		Location:     payload.Location,
		ProfileImage: profileImageKey,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message":      "Profile updated successfully",
		"profileImage": profileImageKey,
	})
}
