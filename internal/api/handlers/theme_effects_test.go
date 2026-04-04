// Copyright (c) 2025-2026, s0up and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/autobrr/qui/internal/database"
	"github.com/autobrr/qui/internal/models"
)

const onePixelPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO5Wn1sAAAAASUVORK5CYII="

func TestThemeEffectsHandler_GetAndUpdate(t *testing.T) {
	t.Parallel()

	handler := newTestThemeEffectsHandler(t)

	getReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/theme-effects?themeId=anime", nil)
	getRec := httptest.NewRecorder()
	handler.Get(getRec, getReq)

	require.Equal(t, http.StatusOK, getRec.Code)

	var settings models.ThemeEffectsSettings
	require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &settings))
	require.Equal(t, models.ThemeEffectsThemeScopeShared, settings.ThemeScope)

	updateBody := bytes.NewBufferString(`{"themeScope":"per-theme","backgroundScope":"per-variation","modeScope":"per-mode","particlesMode":"off","backgroundPosition":"top","backgroundOpacity":37,"overlayStrength":36,"mobileBackground":"disabled"}`)
	updateReq := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/theme-effects?themeId=anime", updateBody)
	updateRec := httptest.NewRecorder()
	handler.Update(updateRec, updateReq)

	require.Equal(t, http.StatusOK, updateRec.Code)
	require.NoError(t, json.Unmarshal(updateRec.Body.Bytes(), &settings))
	require.Equal(t, models.ThemeEffectsThemeScopePerTheme, settings.ThemeScope)
	require.Equal(t, models.ThemeEffectsBackgroundScopePerVariation, settings.BackgroundScope)
	require.Equal(t, models.ThemeEffectsModeScopePerMode, settings.ModeScope)
	require.Equal(t, models.ThemeEffectsParticlesModeOff, settings.ParticlesMode)
	require.Equal(t, 37, settings.BackgroundOpacity)
	require.Equal(t, 36, settings.OverlayStrength)
}

func TestThemeEffectsHandler_UploadAndResolve(t *testing.T) {
	t.Parallel()

	handler := newTestThemeEffectsHandler(t)
	imageBytes, err := base64.StdEncoding.DecodeString(onePixelPNGBase64)
	require.NoError(t, err)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("background", "theme.png")
	require.NoError(t, err)
	_, err = part.Write(imageBytes)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	uploadReq := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/theme-effects/background/shared?themeId=anime", body)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())
	uploadReq = withThemeEffectsSlot(uploadReq, models.ThemeEffectsAssetSlotShared)

	uploadRec := httptest.NewRecorder()
	handler.UploadBackground(uploadRec, uploadReq)
	require.Equal(t, http.StatusOK, uploadRec.Code)

	resolveReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/theme-effects/background/resolved?themeId=anime&variation=sakura&mode=light", nil)
	resolveRec := httptest.NewRecorder()
	handler.GetResolvedBackground(resolveRec, resolveReq)

	require.Equal(t, http.StatusOK, resolveRec.Code)
	require.Equal(t, "image/png", resolveRec.Header().Get("Content-Type"))
	require.NotEmpty(t, resolveRec.Body.Bytes())
}

func TestThemeEffectsHandler_UploadRejectsInvalidSlot(t *testing.T) {
	t.Parallel()

	handler := newTestThemeEffectsHandler(t)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/theme-effects/background/nope?themeId=anime", http.NoBody)
	req = withThemeEffectsSlot(req, "nope")

	rec := httptest.NewRecorder()
	handler.UploadBackground(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func newTestThemeEffectsHandler(t *testing.T) *ThemeEffectsHandler {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "theme-effects.db")
	db, err := database.New(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	store := models.NewThemeEffectsStore(db)
	return NewThemeEffectsHandler(store, t.TempDir())
}

func withThemeEffectsSlot(req *http.Request, slot string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slot", slot)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}
