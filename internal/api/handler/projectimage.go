package handler

import (
	"fmt"
	"io"
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

	data, err := io.ReadAll(file)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	key, err := h.r2Client.UploadImage(r.Context(), data,
		fmt.Sprintf("projects/%d/%s", projectId, imgType),
		httputil.GenerateUUID(),
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

	data, err := io.ReadAll(file)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err)
		return
	}

	key, err := h.r2Client.UploadImage(r.Context(), data,
		fmt.Sprintf("projects/%d", projectId),
		"cover",
	)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err)
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
