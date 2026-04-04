// Copyright (c) 2025-2026, s0up and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package models

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestThemeEffectsStore_GetDefaultsAndUpdate(t *testing.T) {
	t.Parallel()

	sqlDB, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	db := newMockQuerier(sqlDB)
	createThemeEffectsTables(t, db)

	store := NewThemeEffectsStore(db)

	settings, err := store.GetByUserID(t.Context(), 1, "anime")
	require.NoError(t, err)
	require.Equal(t, ThemeEffectsThemeScopeShared, settings.ThemeScope)
	require.Equal(t, ThemeEffectsBackgroundScopeShared, settings.BackgroundScope)
	require.Equal(t, ThemeEffectsModeScopeShared, settings.ModeScope)
	require.Equal(t, ThemeEffectsParticlesModeAuto, settings.ParticlesMode)
	require.Equal(t, ThemeEffectsBackgroundPositionCenter, settings.BackgroundPosition)
	require.Equal(t, ThemeEffectsBackgroundOpacityDefault, settings.BackgroundOpacity)
	require.Equal(t, ThemeEffectsOverlayStrengthDefault, settings.OverlayStrength)
	require.Equal(t, ThemeEffectsMobileBackgroundSame, settings.MobileBackground)

	backgroundOpacity := 34
	overlayStrength := 72
	updated, err := store.Update(t.Context(), 1, "anime", &ThemeEffectsSettingsInput{
		ThemeScope:         ThemeEffectsThemeScopePerTheme,
		BackgroundScope:    ThemeEffectsBackgroundScopePerVariation,
		ModeScope:          ThemeEffectsModeScopePerMode,
		ParticlesMode:      ThemeEffectsParticlesModeOn,
		BackgroundPosition: ThemeEffectsBackgroundPositionBottom,
		BackgroundOpacity:  &backgroundOpacity,
		OverlayStrength:    &overlayStrength,
		MobileBackground:   ThemeEffectsMobileBackgroundDisabled,
	})
	require.NoError(t, err)
	require.Equal(t, ThemeEffectsThemeScopePerTheme, updated.ThemeScope)
	require.Equal(t, ThemeEffectsBackgroundScopePerVariation, updated.BackgroundScope)
	require.Equal(t, ThemeEffectsModeScopePerMode, updated.ModeScope)
	require.Equal(t, ThemeEffectsParticlesModeOn, updated.ParticlesMode)
	require.Equal(t, ThemeEffectsBackgroundPositionBottom, updated.BackgroundPosition)
	require.Equal(t, 34, updated.BackgroundOpacity)
	require.Equal(t, 72, updated.OverlayStrength)
	require.Equal(t, ThemeEffectsMobileBackgroundDisabled, updated.MobileBackground)
}

func TestThemeEffectsStore_ClonesSharedStateWhenSwitchingToPerTheme(t *testing.T) {
	t.Parallel()

	sqlDB, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	db := newMockQuerier(sqlDB)
	createThemeEffectsTables(t, db)

	store := NewThemeEffectsStore(db)
	sharedOpacity := 51
	_, err = store.Update(t.Context(), 1, "anime", &ThemeEffectsSettingsInput{
		BackgroundOpacity: &sharedOpacity,
	})
	require.NoError(t, err)
	_, err = store.SetBackgroundPath(t.Context(), 1, "anime", ThemeEffectsAssetSlotShared, "shared.webp")
	require.NoError(t, err)

	perTheme, err := store.Update(t.Context(), 1, "anime", &ThemeEffectsSettingsInput{
		ThemeScope: ThemeEffectsThemeScopePerTheme,
	})
	require.NoError(t, err)
	require.Equal(t, ThemeEffectsThemeScopePerTheme, perTheme.ThemeScope)
	require.Equal(t, 51, perTheme.BackgroundOpacity)
	require.Equal(t, "shared.webp", perTheme.BackgroundPathForSlot(ThemeEffectsAssetSlotShared))
}

func TestThemeEffectsStore_RejectsStrengthValuesOutsideRange(t *testing.T) {
	t.Parallel()

	sqlDB, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	db := newMockQuerier(sqlDB)
	createThemeEffectsTables(t, db)

	store := NewThemeEffectsStore(db)
	backgroundOpacity := 101
	overlayStrength := 101

	_, err = store.Update(t.Context(), 1, "anime", &ThemeEffectsSettingsInput{
		BackgroundOpacity: &backgroundOpacity,
	})
	require.EqualError(t, err, `invalid backgroundOpacity 101`)

	_, err = store.Update(t.Context(), 1, "anime", &ThemeEffectsSettingsInput{
		OverlayStrength: &overlayStrength,
	})
	require.EqualError(t, err, `invalid overlayStrength 101`)
}

func TestThemeEffectsSettings_ResolveBackgroundPath(t *testing.T) {
	t.Parallel()

	settings := &ThemeEffectsSettings{
		BackgroundScope: ThemeEffectsBackgroundScopePerVariation,
		ModeScope:       ThemeEffectsModeScopePerMode,
	}
	settings.setAssetPaths(map[string]string{
		ThemeEffectsAssetSlotShared: "shared.webp",
		ThemeEffectsAssetSlotDark:   "dark.webp",
		"sakura":                    "sakura.webp",
		"sakura-dark":               "sakura-dark.webp",
	})

	require.Equal(t, "sakura-dark.webp", settings.ResolveBackgroundPath("sakura", ThemeEffectsColorModeDark))
	require.Equal(t, "sakura.webp", settings.ResolveBackgroundPath("sakura", ThemeEffectsColorModeLight))
	require.Equal(t, "dark.webp", settings.ResolveBackgroundPath("ocean", ThemeEffectsColorModeDark))
	require.Equal(t, "shared.webp", settings.ResolveBackgroundPath("", ThemeEffectsColorModeLight))
}

func createThemeEffectsTables(t *testing.T, db *mockQuerier) {
	t.Helper()

	_, err := db.ExecContext(t.Context(), `
		CREATE TABLE theme_effects_preferences (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL UNIQUE,
			theme_scope TEXT NOT NULL DEFAULT 'shared',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(t.Context(), `
		CREATE TABLE theme_effects_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			theme_key TEXT NOT NULL DEFAULT '',
			background_scope TEXT NOT NULL DEFAULT 'shared',
			mode_scope TEXT NOT NULL DEFAULT 'shared',
			particles_mode TEXT NOT NULL DEFAULT 'auto',
			background_position TEXT NOT NULL DEFAULT 'center',
			background_opacity INTEGER NOT NULL DEFAULT 42,
			overlay_strength INTEGER NOT NULL DEFAULT 28,
			mobile_background TEXT NOT NULL DEFAULT 'same',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, theme_key)
		)
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(t.Context(), `
		CREATE TABLE theme_effects_backgrounds (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			theme_key TEXT NOT NULL DEFAULT '',
			slot TEXT NOT NULL,
			path TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, theme_key, slot)
		)
	`)
	require.NoError(t, err)
}
