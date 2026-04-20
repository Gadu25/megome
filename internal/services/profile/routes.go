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
	Profile *types.Profile `json:"profile"`
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
		Profile: profile,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())

	file, handler, err := r.FormFile("profileImage")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("Error handling file: %w", err))
		return
	}

	// 1MB will only allowed
	// TODO compress and convert uploaded files to webp for better user experience.
	if handler.Size > 1<<20 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("file too large (max 1MB)"))
    return
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("failed to read file: %w", err))
		return
	}
	fileType := http.DetectContentType(buffer)

	if fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/webp" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid file type: %w", fileType))
		return
	}

	payload := types.MakeProfilePayload{
		Bio:       r.FormValue("bio"),
		FirstName: r.FormValue("firstName"),
		LastName:  r.FormValue("lastName"),
		Title:     r.FormValue("title"),
		Birthday:  r.FormValue("birthday"),
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

		existing, err := h.profileStore.GetProfile(userID)
		key, err := storage.GenerateKey(fmt.Sprintf("profiles/%d", userID), "avatar", fileType)

		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid file name: %w", err))
			return
		}

		if existing != nil {
			// remove old image
			oldKey := existing.ProfileImage
			err = h.r2Client.DeleteObject(r.Context(), oldKey)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to update image: %w", err))
				return
			}
		}

		err = h.r2Client.UploadFromReader(r.Context(), key, file, header.Size, header.Header.Get("Content-Type"))
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to upload image: %w", err))
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
		Title:        payload.Title,
		Birthday:     payload.Birthday,
		Phone:        payload.Phone,
		Website:      payload.Website,
		Location:     payload.Location,
		ProfileImage: profileImageKey,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	profile, err := h.profileStore.GetProfile(userID)
	
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := ProfileResponse{
		Message: "Profile updated successfully",
		Profile: profile,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
