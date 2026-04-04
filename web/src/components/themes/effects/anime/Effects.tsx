/*
 * Copyright (c) 2025-2026, s0up and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { AnimeThemeBackground } from "@/components/themes/effects/anime/Background"
import { AnimeThemeParticles } from "@/components/themes/effects/anime/Particles"
import { ThemeEffectsOverlay } from "@/components/themes/effects/ThemeEffectsOverlay"
import {
  getResolvedThemeEffectsBackgroundUrl,
  THEME_EFFECTS_SETTINGS_DEFAULTS,
  useThemeEffectsSettings
} from "@/hooks/useThemeEffectsSettings"
import { useTheme } from "@/hooks/useTheme"
import type { AnimeVariation, ThemeEffectsColorMode } from "@/types"

export function AnimeThemeEffects() {
  const { theme, variation, mode } = useTheme()
  const { data: settings = THEME_EFFECTS_SETTINGS_DEFAULTS } = useThemeEffectsSettings(theme, true)

  const activeVariation = (variation ?? "sakura") as AnimeVariation
  const prefersDarkMode = mode === "dark" || (mode === "auto" && document.documentElement.classList.contains("dark"))
  const activeMode = (prefersDarkMode ? "dark" : "light") as ThemeEffectsColorMode
  const hasBackgroundAssets = Object.values(settings.assetSlots ?? {}).some(Boolean)
  const imageUrl = hasBackgroundAssets ? getResolvedThemeEffectsBackgroundUrl(theme, activeVariation, activeMode, settings.updatedAt) : undefined

  return (
    <div className="pointer-events-none fixed inset-0 z-0 overflow-hidden">
      <AnimeThemeBackground
        variation={activeVariation}
        mode={activeMode}
        imageUrl={imageUrl}
        backgroundPosition={settings.backgroundPosition}
        backgroundOpacity={settings.backgroundOpacity}
        mobileBackground={settings.mobileBackground}
      />
      <AnimeThemeParticles variation={activeVariation} particlesMode={settings.particlesMode} />
      <ThemeEffectsOverlay strength={settings.overlayStrength} />
    </div>
  )
}
