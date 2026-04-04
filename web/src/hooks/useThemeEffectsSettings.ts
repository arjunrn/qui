/*
 * Copyright (c) 2025-2026, s0up and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { getThemeById } from "@/config/themes"
import { useHasPremiumAccess } from "@/hooks/useLicense"
import { useTheme } from "@/hooks/useTheme"
import { api } from "@/lib/api"
import { getApiBaseUrl } from "@/lib/base-url"
import { canSwitchToPremiumTheme } from "@/lib/license-entitlement"
import type {
  ThemeEffectsAssetSlot,
  ThemeEffectsColorMode,
  ThemeEffectsSettings,
  ThemeEffectsSettingsInput
} from "@/types"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

const THEME_EFFECTS_TIMEOUT_MS = 8000

export const THEME_EFFECTS_SETTINGS_DEFAULTS: ThemeEffectsSettings = {
  id: 0,
  userId: 0,
  themeScope: "shared",
  backgroundScope: "shared",
  modeScope: "shared",
  particlesMode: "auto",
  backgroundPosition: "center",
  backgroundOpacity: 42,
  overlayStrength: 28,
  mobileBackground: "same",
  assetSlots: {},
  createdAt: "",
  updatedAt: "",
}

export function getThemeEffectsAssetSlots(
  themeId: string,
  backgroundScope: ThemeEffectsSettings["backgroundScope"],
  modeScope: ThemeEffectsSettings["modeScope"]
): ThemeEffectsAssetSlot[] {
  const variations = getThemeById(themeId)?.variations ?? []

  if (backgroundScope === "per-variation" && variations.length > 0 && modeScope === "per-mode") {
    return variations.flatMap((variation) => [`${variation}-light`, `${variation}-dark`] as ThemeEffectsAssetSlot[])
  }
  if (backgroundScope === "per-variation" && variations.length > 0) {
    return variations as ThemeEffectsAssetSlot[]
  }
  if (modeScope === "per-mode") {
    return ["light", "dark"]
  }
  return ["shared"]
}

export function getThemeEffectsAssetSlotLabel(slot: ThemeEffectsAssetSlot): string {
  return slot
    .split("-")
    .map(part => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ")
}

async function fetchThemeEffectsSettings(themeId: string): Promise<ThemeEffectsSettings> {
  const controller = new AbortController()
  const timeoutId = window.setTimeout(() => controller.abort(), THEME_EFFECTS_TIMEOUT_MS)

  try {
    const response = await fetch(`${getApiBaseUrl()}/theme-effects?themeId=${encodeURIComponent(themeId)}`, {
      credentials: "include",
      headers: {
        "X-Requested-With": "XMLHttpRequest",
      },
      signal: controller.signal,
    })

    if (!response.ok) {
      let message = "Failed to load theme effects settings"

      try {
        const body = await response.json() as { error?: string; message?: string }
        if (body.error?.trim()) {
          message = body.error
        } else if (body.message?.trim()) {
          message = body.message
        }
      } catch {
        // Keep generic message for non-JSON errors.
      }

      throw new Error(message)
    }

    return response.json()
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") {
      throw new Error("Timed out loading theme effects settings", { cause: error })
    }

    if (error instanceof Error) {
      throw new Error(error.message, { cause: error })
    }

    throw new Error("Failed to load theme effects settings", { cause: error })
  } finally {
    window.clearTimeout(timeoutId)
  }
}

export function useThemeEffectsSettings(themeId: string, enabled = true) {
  return useQuery<ThemeEffectsSettings>({
    queryKey: ["theme-effects-settings", themeId],
    queryFn: () => fetchThemeEffectsSettings(themeId),
    placeholderData: THEME_EFFECTS_SETTINGS_DEFAULTS,
    staleTime: 60000,
    gcTime: 300000,
    enabled: enabled && themeId.length > 0,
  })
}

export function useUpdateThemeEffectsSettings() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ themeId, data }: { themeId: string; data: ThemeEffectsSettingsInput }) =>
      api.updateThemeEffectsSettings(themeId, data),
    onSuccess: (settings, variables) => {
      queryClient.setQueryData(["theme-effects-settings", variables.themeId], settings)
    },
    onSettled: (_data, _error, variables) => {
      queryClient.invalidateQueries({ queryKey: ["theme-effects-settings", variables.themeId] })
    },
  })
}

export function useUploadThemeEffectsBackground() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ themeId, slot, file }: { themeId: string; slot: ThemeEffectsAssetSlot; file: File }) =>
      api.uploadThemeEffectsBackground(themeId, slot, file),
    onSuccess: (settings, variables) => {
      queryClient.setQueryData(["theme-effects-settings", variables.themeId], settings)
    },
    onSettled: (_data, _error, variables) => {
      queryClient.invalidateQueries({ queryKey: ["theme-effects-settings", variables.themeId] })
    },
  })
}

export function useDeleteThemeEffectsBackground() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ themeId, slot }: { themeId: string; slot: ThemeEffectsAssetSlot }) =>
      api.deleteThemeEffectsBackground(themeId, slot),
    onSettled: (_data, _error, variables) => {
      queryClient.invalidateQueries({ queryKey: ["theme-effects-settings", variables.themeId] })
    },
  })
}

export function getThemeEffectsBackgroundUrl(themeId: string, slot: ThemeEffectsAssetSlot, cacheBuster?: string): string {
  return api.getThemeEffectsBackgroundUrl(themeId, slot, cacheBuster)
}

export function getResolvedThemeEffectsBackgroundUrl(
  themeId: string,
  variation: string | undefined,
  mode: ThemeEffectsColorMode,
  cacheBuster?: string
): string {
  return api.getResolvedThemeEffectsBackgroundUrl(themeId, variation, mode, cacheBuster)
}

export function useCurrentThemeEffectsState() {
  const { theme, currentTheme } = useTheme()
  const { hasPremiumAccess, isLoading, isError } = useHasPremiumAccess()
  const canUseThemeEffects = canSwitchToPremiumTheme({
    hasPremiumAccess,
    isLoading,
    isError,
  })
  const query = useThemeEffectsSettings(theme, canUseThemeEffects)
  const hasCustomBackground = Object.values(query.data?.assetSlots ?? {}).some(Boolean)

  return {
    themeId: theme,
    currentTheme,
    canUseThemeEffects,
    hasCustomBackground,
    ...query,
  }
}
