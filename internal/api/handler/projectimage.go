package handler

import (
	"fmt"
	"megome/internal/domain/project"
	"megome/internal/domain/user"
	"megome/internal/middleware"
	"megome/internal/pkg/httputil"
	"megome/internal/pkg/storage"
	"net/http"

	"github.com/gorilla/mux"
)

type ProjectImageHandler struct {
	projectStore *project.Repository
	userStore    *user.Repository
	r2Client     *storage.R2Client
}

type ImagesResponse struct {
	Message string              `json:"message"`
	Images  []project.ProjectImage `json:"images"`
}

type SingleImageResponse struct {
	Message string              `json:"message"`
	Image   project.ProjectImage `json:"image"`
}

func NewProjectImageHandler(projectStore *project.Repository, userStore *user.Repository, r2Client *storage.R2Client) *ProjectImageHandler {
	return &ProjectImageHandler{projectStore: projectStore, userStore: userStore, r2Client: r2Client}
}

func (h *ProjectImageHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/project/{id}/images", middleware.WithJWTAuth(h.handleGetImages, h.userStore)).Methods("GET")
	router.HandleFunc("/project/{id}/images", middleware.WithJWTAuth(h.handleUploadImage, h.userStore)).Methods("POST")
	router.HandleFunc("/project/{id}/cover", middleware.WithJWTAuth(h.handleSetCover, h.userStore)).Methods("PUT")
	router.HandleFunc("/project-images/{id}", middleware.WithJWTAuth(h.handleDeleteImage, h.userStore)).Methods("DELETE")
}

func (h *ProjectImageHandler) handleGetImages(w http.ResponseWriter, r *http.Request) {
	projectId, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	images, err := h.projectStore.GetProjectImagesByProjectID(projectId)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, ImagesResponse{
		Message: "Images fetched successfully",
		Images:  images,
	})
}

func (h *ProjectImageHandler) handleUploadImage(w http.ResponseWriter, r *http.Request) {
	projectId, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	imgType := r.FormValue("type")

	if imgType != "screenshot" && imgType != "demo" {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid image type"))
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("file required"))
		return
	}
	defer file.Close()

	if handler.Size > 1<<20 {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("file too large (max 1MB)"))
		return
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	fileType := http.DetectContentType(buffer)
	if fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/webp" {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid file type"))
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	key, err := storage.GenerateKey(
		fmt.Sprintf("projects/%d/%s", projectId, imgType),
		httputil.GenerateUUID(),
		fileType,
	)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
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
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	image, err := h.projectStore.AddProjectImage(project.ProjectImage{
		ProjectID: projectId,
		URL:       key,
		Type:      imgType,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, SingleImageResponse{
		Message: "Image uploaded successfully",
		Image:   image,
	})
}

func (h *ProjectImageHandler) handleSetCover(w http.ResponseWriter, r *http.Request) {
	projectId, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("file required"))
		return
	}
	defer file.Close()

	if handler.Size > 1<<20 {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Errorf("file too large"))
		return
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	fileType := http.DetectContentType(buffer)

	file, header, err := r.FormFile("image")
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	key, err := storage.GenerateKey(
		fmt.Sprintf("projects/%d", projectId),
		"cover",
		fileType,
	)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	existingImages, err := h.projectStore.GetProjectImagesByProjectID(projectId)
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
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	image, err := h.projectStore.SetProjectCover(projectId, project.ProjectImage{
		URL: key,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, SingleImageResponse{
		Message: "Cover updated successfully",
		Image:   image,
	})
}

func (h *ProjectImageHandler) handleDeleteImage(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.GetRequestId(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	image, err := h.projectStore.GetProjectImageByID(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if image.URL != "" {
		_ = h.r2Client.DeleteObject(r.Context(), image.URL)
	}

	err = h.projectStore.DeleteProjectImage(id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Image deleted successfully",
	})
}
