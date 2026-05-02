package projectimages

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
	imageStore types.ProjectImageStore
	userStore  types.UserStore
	r2Client   *storage.R2Client
}

type ImagesResponse struct {
	Message string               `json:"message"`
	Images  []types.ProjectImage `json:"images"`
}

type SingleImageResponse struct {
	Message string             `json:"message"`
	Image   types.ProjectImage `json:"image"`
}

func NewHandler(imageStore types.ProjectImageStore, userStore types.UserStore) *Handler {
	return &Handler{imageStore: imageStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/project/{id}/images", auth.WithJWTAuth(h.handleGetImages, h.userStore)).Methods("GET")
	router.HandleFunc("/project/{id}/images", auth.WithJWTAuth(h.handleUploadImage, h.userStore)).Methods("POST")
	router.HandleFunc("/project/{id}/cover", auth.WithJWTAuth(h.handleSetCover, h.userStore)).Methods("PUT")
	router.HandleFunc("/project-images/{id}", auth.WithJWTAuth(h.handleDeleteImage, h.userStore)).Methods("DELETE")
}

func (h *Handler) handleGetImages(w http.ResponseWriter, r *http.Request) {
	projectId, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	images, err := h.imageStore.GetProjectImages(projectId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, ImagesResponse{
		Message: "Images fetched successfully",
		Images:  images,
	})
}

func (h *Handler) handleUploadImage(w http.ResponseWriter, r *http.Request) {
	projectId, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	imgType := r.FormValue("type") // screenshot | demo

	if imgType != "screenshot" && imgType != "demo" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid image type"))
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("file required"))
		return
	}
	defer file.Close()

	// size limit (1MB)
	if handler.Size > 1<<20 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("file too large (max 1MB)"))
		return
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	fileType := http.DetectContentType(buffer)
	if fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/webp" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid file type"))
		return
	}

	// reopen file (same as your profile logic)
	file, header, err := r.FormFile("image")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	key, err := storage.GenerateKey(
		fmt.Sprintf("projects/%d", projectId),
		imgType,
		fileType,
	)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.r2Client.UploadFromReader(
		r.Context(),
		key,
		file,
		header.Size,
		header.Header.Get("Content-Type"),
	)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	image, err := h.imageStore.AddProjectImage(types.ProjectImage{
		ProjectID: projectId,
		URL:       key,
		Type:      imgType,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, SingleImageResponse{
		Message: "Image uploaded successfully",
		Image:   image,
	})
}

func (h *Handler) handleSetCover(w http.ResponseWriter, r *http.Request) {
	projectId, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("file required"))
		return
	}
	defer file.Close()

	if handler.Size > 1<<20 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("file too large"))
		return
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	fileType := http.DetectContentType(buffer)

	file, header, err := r.FormFile("image")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	key, err := storage.GenerateKey(
		fmt.Sprintf("projects/%d", projectId),
		"cover",
		fileType,
	)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// 🔥 delete old cover first
	existingImages, err := h.imageStore.GetProjectImages(projectId)
	if err == nil {
		for _, img := range existingImages {
			if img.Type == "cover" && img.URL != "" {
				_ = h.r2Client.DeleteObject(r.Context(), img.URL)
			}
		}
	}

	err = h.r2Client.UploadFromReader(
		r.Context(),
		key,
		file,
		header.Size,
		header.Header.Get("Content-Type"),
	)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	image, err := h.imageStore.SetProjectCover(projectId, types.ProjectImage{
		URL: key,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, SingleImageResponse{
		Message: "Cover updated successfully",
		Image:   image,
	})
}

func (h *Handler) handleDeleteImage(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetRequestId(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	image, err := h.imageStore.GetProjectImageByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if image.URL != "" {
		_ = h.r2Client.DeleteObject(r.Context(), image.URL)
	}

	err = h.imageStore.DeleteProjectImage(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Image deleted successfully",
	})
}
