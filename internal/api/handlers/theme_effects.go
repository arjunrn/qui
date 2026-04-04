// Copyright (c) 2025-2026, s0up and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/autobrr/qui/internal/models"
)

const themeEffectsBackgroundMaxFormMemory int64 = 5 << 20
const themeEffectsBackgroundMaxBodyBytes int64 = 6 << 20

type ThemeEffectsHandler struct {
	store   *models.ThemeEffectsStore
	dataDir string
}

func NewThemeEffectsHandler(store *models.ThemeEffectsStore, dataDir string) *ThemeEffectsHandler {
	return &ThemeEffectsHandler{
		store:   store,
		dataDir: dataDir,
	}
}

func (h *ThemeEffectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	themeID, ok := h.getThemeID(w, r)
	if !ok {
		return
	}

	settings, err := h.store.GetByUserID(r.Context(), 1, themeID)
	if err != nil {
		log.Error().Err(err).Str("themeId", themeID).Msg("failed to get theme effects settings")
		RespondError(w, http.StatusInternalServerError, "Failed to load theme effects settings")
		return
	}

	RespondJSON(w, http.StatusOK, settings)
}

func (h *ThemeEffectsHandler) Update(w http.ResponseWriter, r *http.Request) {
	themeID, ok := h.getThemeID(w, r)
	if !ok {
		return
	}

	var input models.ThemeEffectsSettingsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Warn().Err(err).Msg("failed to decode theme effects settings request")
		RespondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	settings, err := h.store.Update(r.Context(), 1, themeID, &input)
	if err != nil {
		if isThemeEffectsValidationError(err) {
			RespondError(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Error().Err(err).Str("themeId", themeID).Msg("failed to update theme effects settings")
		RespondError(w, http.StatusInternalServerError, "Failed to update theme effects settings")
		return
	}

	RespondJSON(w, http.StatusOK, settings)
}

func (h *ThemeEffectsHandler) UploadBackground(w http.ResponseWriter, r *http.Request) {
	themeID, ok := h.getThemeID(w, r)
	if !ok {
		return
	}

	slot := chi.URLParam(r, "slot")
	if !models.IsThemeEffectsAssetSlot(slot) {
		RespondError(w, http.StatusBadRequest, "Invalid background slot")
		return
	}

	themeKey, err := h.store.ResolveThemeKey(r.Context(), 1, themeID)
	if err != nil {
		if isThemeEffectsValidationError(err) {
			RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Error().Err(err).Str("themeId", themeID).Msg("failed to resolve theme effects theme key")
		RespondError(w, http.StatusInternalServerError, "Failed to store background image")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, themeEffectsBackgroundMaxBodyBytes)
	if err := r.ParseMultipartForm(themeEffectsBackgroundMaxFormMemory); err != nil {
		if errors.Is(err, multipart.ErrMessageTooLarge) {
			RespondError(w, http.StatusRequestEntityTooLarge, "Upload exceeded 5 MB limit")
			return
		}
		RespondError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	file, _, err := r.FormFile("background")
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Background image is required")
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Warn().Err(err).Str("slot", slot).Msg("failed to read theme effects background upload")
		RespondError(w, http.StatusBadRequest, "Failed to read uploaded image")
		return
	}

	contentType, ext, err := detectThemeEffectsImageType(content)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	oldPath, err := h.store.GetBackgroundPathBySlot(r.Context(), 1, themeID, slot)
	if err != nil {
		log.Error().Err(err).Str("slot", slot).Str("themeId", themeID).Msg("failed to read existing theme effects background path")
		RespondError(w, http.StatusInternalServerError, "Failed to update background image")
		return
	}

	themeDir := "shared"
	if themeKey != "" {
		themeDir = themeKey
	}
	relativePath := filepath.ToSlash(filepath.Join("uploads", "theme-effects", "1", themeDir, slot+ext))
	absolutePath, err := h.backgroundAbsolutePath(relativePath)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to resolve background image path")
		return
	}

	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		log.Error().Err(err).Str("path", absolutePath).Msg("failed to create theme effects background directory")
		RespondError(w, http.StatusInternalServerError, "Failed to store background image")
		return
	}

	if err := os.WriteFile(absolutePath, content, 0o600); err != nil {
		log.Error().Err(err).Str("path", absolutePath).Msg("failed to write theme effects background image")
		RespondError(w, http.StatusInternalServerError, "Failed to store background image")
		return
	}

	settings, err := h.store.SetBackgroundPath(r.Context(), 1, themeID, slot, relativePath)
	if err != nil {
		_ = os.Remove(absolutePath)
		if isThemeEffectsValidationError(err) {
			RespondError(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Error().Err(err).Str("slot", slot).Str("themeId", themeID).Msg("failed to save theme effects background path")
		RespondError(w, http.StatusInternalServerError, "Failed to store background image")
		return
	}

	if oldPath != "" && oldPath != relativePath {
		if oldAbsolutePath, pathErr := h.backgroundAbsolutePath(oldPath); pathErr == nil {
			_ = os.Remove(oldAbsolutePath)
		}
	}

	w.Header().Set("X-Content-Type", contentType)
	RespondJSON(w, http.StatusOK, settings)
}

func (h *ThemeEffectsHandler) DeleteBackground(w http.ResponseWriter, r *http.Request) {
	themeID, ok := h.getThemeID(w, r)
	if !ok {
		return
	}

	slot := chi.URLParam(r, "slot")
	if !models.IsThemeEffectsAssetSlot(slot) {
		RespondError(w, http.StatusBadRequest, "Invalid background slot")
		return
	}

	path, err := h.store.GetBackgroundPathBySlot(r.Context(), 1, themeID, slot)
	if err != nil {
		log.Error().Err(err).Str("slot", slot).Str("themeId", themeID).Msg("failed to read theme effects background path for delete")
		RespondError(w, http.StatusInternalServerError, "Failed to delete background image")
		return
	}

	if _, err := h.store.ClearBackgroundPath(r.Context(), 1, themeID, slot); err != nil {
		if isThemeEffectsValidationError(err) {
			RespondError(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Error().Err(err).Str("slot", slot).Str("themeId", themeID).Msg("failed to clear theme effects background path")
		RespondError(w, http.StatusInternalServerError, "Failed to delete background image")
		return
	}

	if path != "" {
		if absolutePath, err := h.backgroundAbsolutePath(path); err == nil {
			_ = os.Remove(absolutePath)
		}
	}

	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *ThemeEffectsHandler) GetBackground(w http.ResponseWriter, r *http.Request) {
	themeID, ok := h.getThemeID(w, r)
	if !ok {
		return
	}

	slot := chi.URLParam(r, "slot")
	if !models.IsThemeEffectsAssetSlot(slot) {
		RespondError(w, http.StatusBadRequest, "Invalid background slot")
		return
	}

	path, err := h.store.GetBackgroundPathBySlot(r.Context(), 1, themeID, slot)
	if err != nil {
		log.Error().Err(err).Str("slot", slot).Str("themeId", themeID).Msg("failed to read theme effects background path")
		RespondError(w, http.StatusInternalServerError, "Failed to load background image")
		return
	}
	if path == "" {
		RespondError(w, http.StatusNotFound, "Background image not found")
		return
	}

	h.serveBackgroundPath(w, path)
}

func (h *ThemeEffectsHandler) GetResolvedBackground(w http.ResponseWriter, r *http.Request) {
	themeID, ok := h.getThemeID(w, r)
	if !ok {
		return
	}

	variation := strings.TrimSpace(r.URL.Query().Get("variation"))
	mode := strings.TrimSpace(r.URL.Query().Get("mode"))
	if !models.IsThemeEffectsColorMode(mode) {
		RespondError(w, http.StatusBadRequest, "Invalid color mode")
		return
	}

	path, err := h.store.ResolveBackgroundPath(r.Context(), 1, themeID, variation, mode)
	if err != nil {
		if isThemeEffectsValidationError(err) {
			RespondError(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Error().Err(err).Str("themeId", themeID).Msg("failed to load theme effects settings for resolved background")
		RespondError(w, http.StatusInternalServerError, "Failed to load background image")
		return
	}
	if path == "" {
		RespondError(w, http.StatusNotFound, "Background image not found")
		return
	}

	h.serveBackgroundPath(w, path)
}

func (h *ThemeEffectsHandler) serveBackgroundPath(w http.ResponseWriter, relativePath string) {
	absolutePath, err := h.backgroundAbsolutePath(relativePath)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to resolve background image path")
		return
	}

	content, err := os.ReadFile(absolutePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			RespondError(w, http.StatusNotFound, "Background image not found")
			return
		}

		log.Error().Err(err).Str("path", absolutePath).Msg("failed to read theme effects background image")
		RespondError(w, http.StatusInternalServerError, "Failed to load background image")
		return
	}

	contentType, _, err := detectThemeEffectsImageType(content)
	if err != nil {
		log.Warn().Err(err).Str("path", absolutePath).Msg("failed to detect theme effects background content type")
		contentType = http.DetectContentType(content)
	}

	w.Header().Set("Cache-Control", "private, max-age=300")
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (h *ThemeEffectsHandler) backgroundAbsolutePath(relativePath string) (string, error) {
	cleanRelativePath := filepath.Clean(relativePath)
	baseDir := filepath.Clean(h.dataDir)
	absolutePath := filepath.Join(baseDir, cleanRelativePath)
	if !strings.HasPrefix(absolutePath, baseDir+string(filepath.Separator)) && absolutePath != baseDir {
		return "", errors.New("background image path escaped data directory")
	}
	return absolutePath, nil
}

func (h *ThemeEffectsHandler) getThemeID(w http.ResponseWriter, r *http.Request) (string, bool) {
	themeID := strings.TrimSpace(r.URL.Query().Get("themeId"))
	if themeID == "" {
		RespondError(w, http.StatusBadRequest, "themeId is required")
		return "", false
	}

	return themeID, true
}

func isThemeEffectsValidationError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "invalid ")
}

func detectThemeEffectsImageType(content []byte) (string, string, error) {
	switch http.DetectContentType(content) {
	case "image/jpeg":
		return "image/jpeg", ".jpg", nil
	case "image/png":
		return "image/png", ".png", nil
	case "image/webp":
		return "image/webp", ".webp", nil
	default:
		return "", "", errors.New("Only JPEG, PNG, and WebP images are supported")
	}
}
