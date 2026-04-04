// Copyright (c) 2025-2026, s0up and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"strings"
	"time"

	"github.com/autobrr/qui/internal/dbinterface"
)

const (
	ThemeEffectsThemeScopeShared   = "shared"
	ThemeEffectsThemeScopePerTheme = "per-theme"

	ThemeEffectsBackgroundScopeShared       = "shared"
	ThemeEffectsBackgroundScopePerVariation = "per-variation"

	ThemeEffectsModeScopeShared  = "shared"
	ThemeEffectsModeScopePerMode = "per-mode"

	ThemeEffectsParticlesModeOff  = "off"
	ThemeEffectsParticlesModeAuto = "auto"
	ThemeEffectsParticlesModeOn   = "on"

	ThemeEffectsBackgroundPositionTop    = "top"
	ThemeEffectsBackgroundPositionCenter = "center"
	ThemeEffectsBackgroundPositionBottom = "bottom"

	ThemeEffectsBackgroundOpacityMin     = 0
	ThemeEffectsBackgroundOpacityMax     = 100
	ThemeEffectsBackgroundOpacityDefault = 42

	ThemeEffectsOverlayStrengthMin     = 0
	ThemeEffectsOverlayStrengthMax     = 100
	ThemeEffectsOverlayStrengthDefault = 28

	ThemeEffectsMobileBackgroundSame     = "same"
	ThemeEffectsMobileBackgroundDisabled = "disabled"

	ThemeEffectsColorModeLight = "light"
	ThemeEffectsColorModeDark  = "dark"

	ThemeEffectsAssetSlotShared = "shared"
	ThemeEffectsAssetSlotLight  = "light"
	ThemeEffectsAssetSlotDark   = "dark"

	themeEffectsSharedThemeKey = ""
)

var (
	themeEffectsThemeScopes = map[string]struct{}{
		ThemeEffectsThemeScopeShared:   {},
		ThemeEffectsThemeScopePerTheme: {},
	}
	themeEffectsBackgroundScopes = map[string]struct{}{
		ThemeEffectsBackgroundScopeShared:       {},
		ThemeEffectsBackgroundScopePerVariation: {},
	}
	themeEffectsModeScopes = map[string]struct{}{
		ThemeEffectsModeScopeShared:  {},
		ThemeEffectsModeScopePerMode: {},
	}
	themeEffectsParticlesModes = map[string]struct{}{
		ThemeEffectsParticlesModeOff:  {},
		ThemeEffectsParticlesModeAuto: {},
		ThemeEffectsParticlesModeOn:   {},
	}
	themeEffectsBackgroundPositions = map[string]struct{}{
		ThemeEffectsBackgroundPositionTop:    {},
		ThemeEffectsBackgroundPositionCenter: {},
		ThemeEffectsBackgroundPositionBottom: {},
	}
	themeEffectsMobileBackgroundModes = map[string]struct{}{
		ThemeEffectsMobileBackgroundSame:     {},
		ThemeEffectsMobileBackgroundDisabled: {},
	}
	themeEffectsColorModes = map[string]struct{}{
		ThemeEffectsColorModeLight: {},
		ThemeEffectsColorModeDark:  {},
	}
	themeEffectsIdentifierPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

type ThemeEffectsSettings struct {
	ID                 int             `json:"id"`
	UserID             int             `json:"userId"`
	ThemeScope         string          `json:"themeScope"`
	BackgroundScope    string          `json:"backgroundScope"`
	ModeScope          string          `json:"modeScope"`
	ParticlesMode      string          `json:"particlesMode"`
	BackgroundPosition string          `json:"backgroundPosition"`
	BackgroundOpacity  int             `json:"backgroundOpacity"`
	OverlayStrength    int             `json:"overlayStrength"`
	MobileBackground   string          `json:"mobileBackground"`
	AssetSlots         map[string]bool `json:"assetSlots"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`

	themeKey   string
	assetPaths map[string]string
}

type ThemeEffectsSettingsInput struct {
	ThemeScope         string `json:"themeScope,omitempty"`
	BackgroundScope    string `json:"backgroundScope,omitempty"`
	ModeScope          string `json:"modeScope,omitempty"`
	ParticlesMode      string `json:"particlesMode,omitempty"`
	BackgroundPosition string `json:"backgroundPosition,omitempty"`
	BackgroundOpacity  *int   `json:"backgroundOpacity,omitempty"`
	OverlayStrength    *int   `json:"overlayStrength,omitempty"`
	MobileBackground   string `json:"mobileBackground,omitempty"`
}

type ThemeEffectsStore struct {
	db dbinterface.Querier
}

func NewThemeEffectsStore(db dbinterface.Querier) *ThemeEffectsStore {
	return &ThemeEffectsStore{db: db}
}

func (s *ThemeEffectsStore) GetByUserID(ctx context.Context, userID int, themeID string) (*ThemeEffectsSettings, error) {
	preferences, err := s.getOrCreatePreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	themeKey, err := resolveThemeEffectsThemeKey(preferences.ThemeScope, themeID)
	if err != nil {
		return nil, err
	}

	settings, assetPaths, err := s.getOrCreateThemeState(ctx, userID, themeKey, nil)
	if err != nil {
		return nil, err
	}

	settings.ThemeScope = preferences.ThemeScope
	settings.themeKey = themeKey
	settings.setAssetPaths(assetPaths)
	return settings, nil
}

func (s *ThemeEffectsStore) Update(ctx context.Context, userID int, themeID string, input *ThemeEffectsSettingsInput) (*ThemeEffectsSettings, error) {
	if input == nil {
		return nil, errors.New("theme effects settings input is nil")
	}

	preferences, err := s.getOrCreatePreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	currentThemeKey, err := resolveThemeEffectsThemeKey(preferences.ThemeScope, themeID)
	if err != nil {
		return nil, err
	}

	currentSettings, currentAssetPaths, err := s.getOrCreateThemeState(ctx, userID, currentThemeKey, nil)
	if err != nil {
		return nil, err
	}

	nextThemeScope := preferences.ThemeScope
	if input.ThemeScope != "" {
		if err := validateValue("themeScope", input.ThemeScope, themeEffectsThemeScopes); err != nil {
			return nil, err
		}
		nextThemeScope = input.ThemeScope
	}

	targetThemeKey, err := resolveThemeEffectsThemeKey(nextThemeScope, themeID)
	if err != nil {
		return nil, err
	}

	settings := currentSettings
	assetPaths := currentAssetPaths
	if targetThemeKey != currentThemeKey {
		settings, assetPaths, err = s.getOrCreateThemeState(ctx, userID, targetThemeKey, currentSettings)
		if err != nil {
			return nil, err
		}
		if len(assetPaths) == 0 && len(currentAssetPaths) > 0 {
			if err := s.copyBackgroundPaths(ctx, userID, targetThemeKey, currentAssetPaths); err != nil {
				return nil, err
			}
			assetPaths = maps.Clone(currentAssetPaths)
		}
	}

	if input.BackgroundScope != "" {
		if err := validateValue("backgroundScope", input.BackgroundScope, themeEffectsBackgroundScopes); err != nil {
			return nil, err
		}
		settings.BackgroundScope = input.BackgroundScope
	}
	if input.ModeScope != "" {
		if err := validateValue("modeScope", input.ModeScope, themeEffectsModeScopes); err != nil {
			return nil, err
		}
		settings.ModeScope = input.ModeScope
	}
	if input.ParticlesMode != "" {
		if err := validateValue("particlesMode", input.ParticlesMode, themeEffectsParticlesModes); err != nil {
			return nil, err
		}
		settings.ParticlesMode = input.ParticlesMode
	}
	if input.BackgroundPosition != "" {
		if err := validateValue("backgroundPosition", input.BackgroundPosition, themeEffectsBackgroundPositions); err != nil {
			return nil, err
		}
		settings.BackgroundPosition = input.BackgroundPosition
	}
	if input.BackgroundOpacity != nil {
		if err := validateBackgroundOpacity(*input.BackgroundOpacity); err != nil {
			return nil, err
		}
		settings.BackgroundOpacity = *input.BackgroundOpacity
	}
	if input.OverlayStrength != nil {
		if err := validateOverlayStrength(*input.OverlayStrength); err != nil {
			return nil, err
		}
		settings.OverlayStrength = *input.OverlayStrength
	}
	if input.MobileBackground != "" {
		if err := validateValue("mobileBackground", input.MobileBackground, themeEffectsMobileBackgroundModes); err != nil {
			return nil, err
		}
		settings.MobileBackground = input.MobileBackground
	}

	if err := s.savePreferences(ctx, userID, nextThemeScope); err != nil {
		return nil, err
	}
	if err := s.saveThemeSettings(ctx, userID, targetThemeKey, settings); err != nil {
		return nil, err
	}

	settings.ThemeScope = nextThemeScope
	settings.themeKey = targetThemeKey
	settings.setAssetPaths(assetPaths)
	return s.GetByUserID(ctx, userID, themeID)
}

func (s *ThemeEffectsStore) ResolveThemeKey(ctx context.Context, userID int, themeID string) (string, error) {
	preferences, err := s.getOrCreatePreferences(ctx, userID)
	if err != nil {
		return "", err
	}

	return resolveThemeEffectsThemeKey(preferences.ThemeScope, themeID)
}

func (s *ThemeEffectsStore) SetBackgroundPath(ctx context.Context, userID int, themeID string, slot string, path string) (*ThemeEffectsSettings, error) {
	if err := validateThemeEffectsAssetSlot(slot); err != nil {
		return nil, err
	}

	settings, err := s.GetByUserID(ctx, userID, themeID)
	if err != nil {
		return nil, err
	}

	if path == "" {
		_, err = s.db.ExecContext(ctx, `
			DELETE FROM theme_effects_backgrounds
			WHERE user_id = ? AND theme_key = ? AND slot = ?
		`, userID, settings.themeKey, slot)
	} else {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO theme_effects_backgrounds (user_id, theme_key, slot, path)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(user_id, theme_key, slot) DO UPDATE SET
				path = excluded.path,
				updated_at = CURRENT_TIMESTAMP
		`, userID, settings.themeKey, slot, path)
	}
	if err != nil {
		return nil, err
	}

	return s.GetByUserID(ctx, userID, themeID)
}

func (s *ThemeEffectsStore) ClearBackgroundPath(ctx context.Context, userID int, themeID string, slot string) (*ThemeEffectsSettings, error) {
	return s.SetBackgroundPath(ctx, userID, themeID, slot, "")
}

func (s *ThemeEffectsStore) GetBackgroundPathBySlot(ctx context.Context, userID int, themeID string, slot string) (string, error) {
	if err := validateThemeEffectsAssetSlot(slot); err != nil {
		return "", err
	}

	settings, err := s.GetByUserID(ctx, userID, themeID)
	if err != nil {
		return "", err
	}

	return settings.BackgroundPathForSlot(slot), nil
}

func (s *ThemeEffectsStore) ResolveBackgroundPath(ctx context.Context, userID int, themeID string, variation string, mode string) (string, error) {
	settings, err := s.GetByUserID(ctx, userID, themeID)
	if err != nil {
		return "", err
	}

	return settings.ResolveBackgroundPath(variation, mode), nil
}

func (s *ThemeEffectsStore) getOrCreatePreferences(ctx context.Context, userID int) (*themeEffectsPreferencesRow, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, theme_scope, created_at, updated_at
		FROM theme_effects_preferences
		WHERE user_id = ?
	`, userID)

	preferences := &themeEffectsPreferencesRow{}
	err := row.Scan(
		&preferences.ID,
		&preferences.UserID,
		&preferences.ThemeScope,
		&preferences.CreatedAt,
		&preferences.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		if _, execErr := s.db.ExecContext(ctx, `
			INSERT INTO theme_effects_preferences (user_id, theme_scope)
			VALUES (?, ?)
		`, userID, ThemeEffectsThemeScopeShared); execErr != nil {
			return nil, execErr
		}

		preferences.ThemeScope = ThemeEffectsThemeScopeShared
		return s.getOrCreatePreferences(ctx, userID)
	}
	if err != nil {
		return nil, err
	}

	return preferences, nil
}

func (s *ThemeEffectsStore) savePreferences(ctx context.Context, userID int, themeScope string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE theme_effects_preferences
		SET theme_scope = ?, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
	`, themeScope, userID)
	return err
}

func (s *ThemeEffectsStore) getOrCreateThemeState(ctx context.Context, userID int, themeKey string, seed *ThemeEffectsSettings) (*ThemeEffectsSettings, map[string]string, error) {
	settings, err := s.loadThemeSettings(ctx, userID, themeKey)
	if errors.Is(err, sql.ErrNoRows) {
		if err := s.createThemeSettings(ctx, userID, themeKey, seed); err != nil {
			return nil, nil, err
		}
		settings, err = s.loadThemeSettings(ctx, userID, themeKey)
	}
	if err != nil {
		return nil, nil, err
	}

	assetPaths, err := s.loadBackgroundPaths(ctx, userID, themeKey)
	if err != nil {
		return nil, nil, err
	}

	settings.themeKey = themeKey
	settings.setAssetPaths(assetPaths)
	return settings, assetPaths, nil
}

func (s *ThemeEffectsStore) loadThemeSettings(ctx context.Context, userID int, themeKey string) (*ThemeEffectsSettings, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, background_scope, mode_scope, particles_mode,
		       background_position, background_opacity, overlay_strength, mobile_background,
		       created_at, updated_at
		FROM theme_effects_settings
		WHERE user_id = ? AND theme_key = ?
	`, userID, themeKey)

	settings := &ThemeEffectsSettings{}
	err := row.Scan(
		&settings.ID,
		&settings.UserID,
		&settings.BackgroundScope,
		&settings.ModeScope,
		&settings.ParticlesMode,
		&settings.BackgroundPosition,
		&settings.BackgroundOpacity,
		&settings.OverlayStrength,
		&settings.MobileBackground,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (s *ThemeEffectsStore) createThemeSettings(ctx context.Context, userID int, themeKey string, seed *ThemeEffectsSettings) error {
	backgroundScope := ThemeEffectsBackgroundScopeShared
	modeScope := ThemeEffectsModeScopeShared
	particlesMode := ThemeEffectsParticlesModeAuto
	backgroundPosition := ThemeEffectsBackgroundPositionCenter
	backgroundOpacity := ThemeEffectsBackgroundOpacityDefault
	overlayStrength := ThemeEffectsOverlayStrengthDefault
	mobileBackground := ThemeEffectsMobileBackgroundSame

	if seed != nil {
		backgroundScope = seed.BackgroundScope
		modeScope = seed.ModeScope
		particlesMode = seed.ParticlesMode
		backgroundPosition = seed.BackgroundPosition
		backgroundOpacity = seed.BackgroundOpacity
		overlayStrength = seed.OverlayStrength
		mobileBackground = seed.MobileBackground
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO theme_effects_settings (
			user_id, theme_key, background_scope, mode_scope, particles_mode,
			background_position, background_opacity, overlay_strength, mobile_background
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		userID,
		themeKey,
		backgroundScope,
		modeScope,
		particlesMode,
		backgroundPosition,
		backgroundOpacity,
		overlayStrength,
		mobileBackground,
	)
	return err
}

func (s *ThemeEffectsStore) saveThemeSettings(ctx context.Context, userID int, themeKey string, settings *ThemeEffectsSettings) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE theme_effects_settings
		SET background_scope = ?, mode_scope = ?, particles_mode = ?,
		    background_position = ?, background_opacity = ?, overlay_strength = ?,
		    mobile_background = ?, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND theme_key = ?
	`,
		settings.BackgroundScope,
		settings.ModeScope,
		settings.ParticlesMode,
		settings.BackgroundPosition,
		settings.BackgroundOpacity,
		settings.OverlayStrength,
		settings.MobileBackground,
		userID,
		themeKey,
	)
	return err
}

func (s *ThemeEffectsStore) loadBackgroundPaths(ctx context.Context, userID int, themeKey string) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT slot, path
		FROM theme_effects_backgrounds
		WHERE user_id = ? AND theme_key = ?
	`, userID, themeKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	paths := map[string]string{}
	for rows.Next() {
		var slot string
		var path string
		if err := rows.Scan(&slot, &path); err != nil {
			return nil, err
		}
		paths[slot] = path
	}

	return paths, rows.Err()
}

func (s *ThemeEffectsStore) copyBackgroundPaths(ctx context.Context, userID int, themeKey string, assetPaths map[string]string) error {
	for slot, path := range assetPaths {
		if path == "" {
			continue
		}
		if _, err := s.db.ExecContext(ctx, `
			INSERT INTO theme_effects_backgrounds (user_id, theme_key, slot, path)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(user_id, theme_key, slot) DO UPDATE SET
				path = excluded.path,
				updated_at = CURRENT_TIMESTAMP
		`, userID, themeKey, slot, path); err != nil {
			return err
		}
	}

	return nil
}

func (s *ThemeEffectsSettings) ResolveBackgroundPath(variation string, mode string) string {
	for _, slot := range s.resolveSlotCandidates(variation, mode) {
		if path := s.BackgroundPathForSlot(slot); path != "" {
			return path
		}
	}

	return ""
}

func (s *ThemeEffectsSettings) BackgroundPathForSlot(slot string) string {
	if s.assetPaths == nil {
		return ""
	}
	return s.assetPaths[slot]
}

func (s *ThemeEffectsSettings) resolveSlotCandidates(variation string, mode string) []string {
	variation = normalizeThemeEffectsIdentifier(variation)
	mode = normalizeThemeEffectsMode(mode)

	modeSlot := ThemeEffectsAssetSlotLight
	if mode == ThemeEffectsColorModeDark {
		modeSlot = ThemeEffectsAssetSlotDark
	}

	variationSlot := variation
	variationModeSlot := variation
	if variation != "" {
		variationModeSlot = variation + "-" + mode
	}

	switch {
	case s.BackgroundScope == ThemeEffectsBackgroundScopePerVariation && variation != "" && s.ModeScope == ThemeEffectsModeScopePerMode:
		return []string{variationModeSlot, variationSlot, modeSlot, ThemeEffectsAssetSlotShared}
	case s.BackgroundScope == ThemeEffectsBackgroundScopePerVariation && variation != "":
		return []string{variationSlot, ThemeEffectsAssetSlotShared}
	case s.ModeScope == ThemeEffectsModeScopePerMode:
		return []string{modeSlot, ThemeEffectsAssetSlotShared}
	default:
		return []string{ThemeEffectsAssetSlotShared}
	}
}

func (s *ThemeEffectsSettings) setAssetPaths(paths map[string]string) {
	s.assetPaths = maps.Clone(paths)
	s.AssetSlots = make(map[string]bool, len(paths))
	for slot, path := range paths {
		s.AssetSlots[slot] = path != ""
	}
}

func IsThemeEffectsAssetSlot(slot string) bool {
	if slot == ThemeEffectsAssetSlotShared || slot == ThemeEffectsAssetSlotLight || slot == ThemeEffectsAssetSlotDark {
		return true
	}

	identifier := normalizeThemeEffectsIdentifier(slot)
	if identifier == "" {
		return false
	}

	if strings.HasSuffix(identifier, "-light") {
		return normalizeThemeEffectsIdentifier(strings.TrimSuffix(identifier, "-light")) != ""
	}
	if strings.HasSuffix(identifier, "-dark") {
		return normalizeThemeEffectsIdentifier(strings.TrimSuffix(identifier, "-dark")) != ""
	}

	return identifier == slot
}

func IsThemeEffectsColorMode(mode string) bool {
	_, ok := themeEffectsColorModes[normalizeThemeEffectsMode(mode)]
	return ok
}

func validateThemeEffectsAssetSlot(slot string) error {
	if IsThemeEffectsAssetSlot(slot) {
		return nil
	}
	return fmt.Errorf("invalid background slot %q", slot)
}

func validateBackgroundOpacity(value int) error {
	if value < ThemeEffectsBackgroundOpacityMin || value > ThemeEffectsBackgroundOpacityMax {
		return fmt.Errorf("invalid backgroundOpacity %d", value)
	}

	return nil
}

func validateOverlayStrength(value int) error {
	if value < ThemeEffectsOverlayStrengthMin || value > ThemeEffectsOverlayStrengthMax {
		return fmt.Errorf("invalid overlayStrength %d", value)
	}

	return nil
}

func validateValue(name string, value string, allowed map[string]struct{}) error {
	if _, ok := allowed[value]; ok {
		return nil
	}
	return fmt.Errorf("invalid %s %q", name, value)
}

func resolveThemeEffectsThemeKey(themeScope string, themeID string) (string, error) {
	if themeScope == ThemeEffectsThemeScopeShared {
		return themeEffectsSharedThemeKey, nil
	}

	normalizedThemeID := normalizeThemeEffectsIdentifier(themeID)
	if normalizedThemeID == "" {
		return "", errors.New("themeId is required for per-theme theme effects")
	}

	return normalizedThemeID, nil
}

func normalizeThemeEffectsMode(mode string) string {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if _, ok := themeEffectsColorModes[mode]; ok {
		return mode
	}
	return ThemeEffectsColorModeLight
}

func normalizeThemeEffectsIdentifier(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if !themeEffectsIdentifierPattern.MatchString(value) {
		return ""
	}
	return value
}

type themeEffectsPreferencesRow struct {
	ID         int
	UserID     int
	ThemeScope string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
