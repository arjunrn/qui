/*
 * Copyright (c) 2025-2026, s0up and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { AnimeThemeEffects } from "@/components/themes/effects/anime/Effects"
import { ThemeBackgroundImage } from "@/components/themes/effects/ThemeBackgroundImage"
import { ThemeEffectsOverlay } from "@/components/themes/effects/ThemeEffectsOverlay"
import { ThemeParticles } from "@/components/themes/effects/ThemeParticles"
import {
  getResolvedThemeEffectsBackgroundUrl,
  THEME_EFFECTS_SETTINGS_DEFAULTS,
  useThemeEffectsSettings
} from "@/hooks/useThemeEffectsSettings"
import { useHasPremiumAccess } from "@/hooks/useLicense"
import { useTheme } from "@/hooks/useTheme"
import { canSwitchToPremiumTheme } from "@/lib/license-entitlement"
import type { ThemeEffectsColorMode } from "@/types"
import { useEffect } from "react"

export function PremiumThemeEffects() {
  const { theme, variation, mode } = useTheme()
  const { hasPremiumAccess, isLoading, isError } = useHasPremiumAccess()
  const canUseThemeEffects = canSwitchToPremiumTheme({
    hasPremiumAccess,
    isLoading,
    isError,
  })
  const { data: settings = THEME_EFFECTS_SETTINGS_DEFAULTS } = useThemeEffectsSettings(theme, canUseThemeEffects)
  const prefersDarkMode = mode === "dark" || (mode === "auto" && document.documentElement.classList.contains("dark"))
  const activeMode = (prefersDarkMode ? "dark" : "light") as ThemeEffectsColorMode
  const hasBackgroundAssets = Object.values(settings.assetSlots ?? {}).some(Boolean)
  const imageUrl = canUseThemeEffects && hasBackgroundAssets ? getResolvedThemeEffectsBackgroundUrl(theme, variation ?? undefined, activeMode, settings.updatedAt) : undefined
  const hasCustomBackground = Boolean(imageUrl)

  useEffect(() => {
    const root = document.documentElement

    if (canUseThemeEffects && hasCustomBackground) {
      root.dataset.themeEffectsBackground = "true"
    } else {
      delete root.dataset.themeEffectsBackground
    }

    if (canUseThemeEffects && hasCustomBackground && theme !== "anime") {
      root.dataset.themeEffectsSurface = "generic"
    } else {
      delete root.dataset.themeEffectsSurface
    }

    return () => {
      delete root.dataset.themeEffectsBackground
      delete root.dataset.themeEffectsSurface
    }
  }, [canUseThemeEffects, hasCustomBackground, theme])

  if (!canUseThemeEffects) {
    return null
  }

  if (theme === "anime") {
    return <AnimeThemeEffects />
  }

  if (!hasCustomBackground) {
    return null
  }

  return (
    <div className="pointer-events-none fixed inset-0 z-0 overflow-hidden">
      <ThemeBackgroundImage
        imageUrl={imageUrl}
        backgroundPosition={settings.backgroundPosition}
        backgroundOpacity={settings.backgroundOpacity}
        mobileBackground={settings.mobileBackground}
      />
      <ThemeParticles themeId={theme} mode={activeMode} particlesMode={settings.particlesMode} />
      <ThemeEffectsOverlay strength={settings.overlayStrength} />
    </div>
  )
}
