/*
 * Copyright (c) 2025-2026, s0up and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { getThemeById } from "@/config/themes"
import { useHasPremiumAccess } from "@/hooks/useLicense"
import { useTheme } from "@/hooks/useTheme"
import {
  THEME_EFFECTS_SETTINGS_DEFAULTS,
  getThemeEffectsAssetSlotLabel,
  getThemeEffectsAssetSlots,
  getThemeEffectsBackgroundUrl,
  useDeleteThemeEffectsBackground,
  useThemeEffectsSettings,
  useUpdateThemeEffectsSettings,
  useUploadThemeEffectsBackground
} from "@/hooks/useThemeEffectsSettings"
import { canSwitchToPremiumTheme } from "@/lib/license-entitlement"
import type { ThemeEffectsAssetSlot, ThemeEffectsSettingsInput } from "@/types"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Slider } from "@/components/ui/slider"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select"
import { Image, Loader2, Trash2, UploadCloud } from "lucide-react"
import { useEffect, useId, useMemo, useState } from "react"
import { toast } from "sonner"

const selectGroups = {
  themeScope: [
    { value: "shared", label: "One setup for all themes" },
    { value: "per-theme", label: "Customize each theme separately" },
  ],
  backgroundScope: [
    { value: "shared", label: "One background for this theme" },
    { value: "per-variation", label: "Separate backgrounds per variation" },
  ],
  modeScope: [
    { value: "shared", label: "Reuse one image in light and dark" },
    { value: "per-mode", label: "Separate light and dark assets" },
  ],
  particlesMode: [
    { value: "auto", label: "Automatic" },
    { value: "on", label: "Always on" },
    { value: "off", label: "Off" },
  ],
  backgroundPosition: [
    { value: "top", label: "Top" },
    { value: "center", label: "Center" },
    { value: "bottom", label: "Bottom" },
  ],
  mobileBackground: [
    { value: "same", label: "Keep image on mobile" },
    { value: "disabled", label: "Hide uploaded image on mobile" },
  ],
} as const

type DraftState = Required<ThemeEffectsSettingsInput>

function buildDraft(settings = THEME_EFFECTS_SETTINGS_DEFAULTS): DraftState {
  return {
    themeScope: settings.themeScope,
    backgroundScope: settings.backgroundScope,
    modeScope: settings.modeScope,
    particlesMode: settings.particlesMode,
    backgroundPosition: settings.backgroundPosition,
    backgroundOpacity: settings.backgroundOpacity,
    overlayStrength: settings.overlayStrength,
    mobileBackground: settings.mobileBackground,
  }
}

interface ThemeEffectsSettingsProps {
  themeId?: string
  compact?: boolean
}

function getErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

export function ThemeEffectsSettings({ themeId, compact = false }: ThemeEffectsSettingsProps) {
  const fileInputPrefix = useId()
  const { theme: activeThemeId } = useTheme()
  const { hasPremiumAccess, isLoading: isPremiumAccessLoading, isError: isPremiumAccessError } = useHasPremiumAccess()
  const effectiveThemeId = themeId ?? activeThemeId
  const activeTheme = getThemeById(effectiveThemeId)
  const canUseThemeEffects = canSwitchToPremiumTheme({
    hasPremiumAccess,
    isLoading: isPremiumAccessLoading,
    isError: isPremiumAccessError,
  })
  const supportsVariations = (activeTheme?.variations?.length ?? 0) > 0
  const { data: settings = THEME_EFFECTS_SETTINGS_DEFAULTS, isLoading } = useThemeEffectsSettings(effectiveThemeId, canUseThemeEffects)
  const updateMutation = useUpdateThemeEffectsSettings()
  const uploadMutation = useUploadThemeEffectsBackground()
  const deleteMutation = useDeleteThemeEffectsBackground()
  const [draft, setDraft] = useState<DraftState>(() => buildDraft())

  useEffect(() => {
    setDraft(buildDraft(settings))
  }, [settings])

  useEffect(() => {
    if (!supportsVariations && draft.backgroundScope !== "shared") {
      setDraft(prev => ({ ...prev, backgroundScope: "shared" }))
    }
  }, [draft.backgroundScope, supportsVariations])

  const slotList = useMemo(() => {
    return getThemeEffectsAssetSlots(
      effectiveThemeId,
      supportsVariations ? draft.backgroundScope : "shared",
      draft.modeScope
    )
  }, [draft.backgroundScope, draft.modeScope, effectiveThemeId, supportsVariations])

  const hasPendingMutation = updateMutation.isPending || uploadMutation.isPending || deleteMutation.isPending
  const isDirty = JSON.stringify(draft) !== JSON.stringify(buildDraft(settings))

  if (!activeTheme) {
    return null
  }

  if (!canUseThemeEffects) {
    return (
      <div className="rounded-lg border border-dashed p-4 text-sm text-muted-foreground">
        Theme effects are available on all themes, but require premium access.
      </div>
    )
  }

  const normalizedDraft: DraftState = {
    ...draft,
    backgroundScope: supportsVariations ? draft.backgroundScope : "shared",
    particlesMode: draft.particlesMode,
  }
  const isSharedScope = settings.themeScope === "shared"

  const handleSave = async () => {
    try {
      await updateMutation.mutateAsync({
        themeId: effectiveThemeId,
        data: normalizedDraft,
      })
      toast.success("Theme effects saved")
    } catch (error) {
      console.error("[ThemeEffectsSettings] save failed", error)
      toast.error(getErrorMessage(error, "Failed to save theme effects"))
    }
  }

  const handleUpload = async (slot: ThemeEffectsAssetSlot, file?: File) => {
    if (!file) {
      return
    }

    try {
      await uploadMutation.mutateAsync({ themeId: effectiveThemeId, slot, file })
      toast.success(`${getThemeEffectsAssetSlotLabel(slot)} background uploaded`)
    } catch (error) {
      console.error("[ThemeEffectsSettings] upload failed", error)
      toast.error(getErrorMessage(error, "Failed to upload background image"))
    }
  }

  const handleDelete = async (slot: ThemeEffectsAssetSlot) => {
    try {
      await deleteMutation.mutateAsync({ themeId: effectiveThemeId, slot })
      toast.success(`${getThemeEffectsAssetSlotLabel(slot)} background removed`)
    } catch (error) {
      console.error("[ThemeEffectsSettings] delete failed", error)
      toast.error(getErrorMessage(error, "Failed to remove background image"))
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        Loading theme effects...
      </div>
    )
  }

  return (
    <div className={`space-y-6 ${compact ? "max-h-[70vh] overflow-y-auto pr-1" : ""}`}>
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Badge variant="secondary">{isSharedScope ? "All Themes" : activeTheme.name}</Badge>
          <Badge variant="outline">
            {isSharedScope ? "Shared across all themes" : "Per-theme setup"}
          </Badge>
        </div>
        <p className="text-sm text-muted-foreground">
          Upload custom backgrounds, tune opacity and readability, and keep one setup shared across all themes or split it per theme.
        </p>
        {isSharedScope && (
          <p className="text-xs text-muted-foreground">
            You are editing the shared setup. Theme-specific visuals still render when the active theme supports them.
          </p>
        )}
        <p className="text-xs text-muted-foreground">
          Recommended: 2560x1440 or larger, WebP/JPEG/PNG, under 5 MB.
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <div className="space-y-2">
          <Label>Theme Scope</Label>
          <Select value={draft.themeScope} onValueChange={(value) => setDraft(prev => ({ ...prev, themeScope: value as DraftState["themeScope"] }))}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {selectGroups.themeScope.map(option => (
                <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label>Light / Dark Scope</Label>
          <Select value={draft.modeScope} onValueChange={(value) => setDraft(prev => ({ ...prev, modeScope: value as DraftState["modeScope"] }))}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {selectGroups.modeScope.map(option => (
                <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {supportsVariations && (
          <div className="space-y-2">
            <Label>Background Scope</Label>
            <Select value={draft.backgroundScope} onValueChange={(value) => setDraft(prev => ({ ...prev, backgroundScope: value as DraftState["backgroundScope"] }))}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {selectGroups.backgroundScope.map(option => (
                  <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}

        <div className="space-y-2">
          <Label>Image Position</Label>
          <Select value={draft.backgroundPosition} onValueChange={(value) => setDraft(prev => ({ ...prev, backgroundPosition: value as DraftState["backgroundPosition"] }))}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {selectGroups.backgroundPosition.map(option => (
                <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label>Particles</Label>
          <Select value={draft.particlesMode} onValueChange={(value) => setDraft(prev => ({ ...prev, particlesMode: value as DraftState["particlesMode"] }))}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {selectGroups.particlesMode.map(option => (
                <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label>Mobile Behavior</Label>
          <Select value={draft.mobileBackground} onValueChange={(value) => setDraft(prev => ({ ...prev, mobileBackground: value as DraftState["mobileBackground"] }))}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {selectGroups.mobileBackground.map(option => (
                <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <div className="flex items-center justify-between gap-3">
            <Label htmlFor={`${fileInputPrefix}-background-opacity`}>Background Opacity</Label>
            <span className="text-sm text-muted-foreground">{draft.backgroundOpacity}%</span>
          </div>
          <Slider
            id={`${fileInputPrefix}-background-opacity`}
            min={0}
            max={100}
            step={1}
            value={[draft.backgroundOpacity]}
            onValueChange={(value) => setDraft(prev => ({ ...prev, backgroundOpacity: value[0] ?? prev.backgroundOpacity }))}
          />
          <p className="text-xs text-muted-foreground">
            Recommended: 25-55. Lower values blend art into the UI and help dense screens stay readable.
          </p>
        </div>

        <div className="space-y-2">
          <div className="flex items-center justify-between gap-3">
            <Label htmlFor={`${fileInputPrefix}-overlay-strength`}>Overlay Strength</Label>
            <span className="text-sm text-muted-foreground">{draft.overlayStrength}%</span>
          </div>
          <Slider
            id={`${fileInputPrefix}-overlay-strength`}
            min={0}
            max={100}
            step={1}
            value={[draft.overlayStrength]}
            onValueChange={(value) => setDraft(prev => ({ ...prev, overlayStrength: value[0] ?? prev.overlayStrength }))}
          />
          <p className="text-xs text-muted-foreground">
            Recommended: 15-45 for most backgrounds. Use this after you tune image opacity.
          </p>
        </div>
      </div>

      <div className="space-y-3">
        <div className="flex items-center justify-between gap-3">
          <div>
            <h4 className="text-sm font-medium">Uploaded Backgrounds</h4>
            <p className="text-xs text-muted-foreground">
              {isSharedScope ? "Active slots follow the scope choices above for the shared theme setup." : `Active slots follow the scope choices above for ${activeTheme.name}.`}
            </p>
          </div>
          <Button type="button" onClick={() => void handleSave()} disabled={!isDirty || hasPendingMutation}>
            {updateMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Save Settings
          </Button>
        </div>

        <div className="grid gap-3 md:grid-cols-2">
          {slotList.map((slot) => {
            const hasImage = Boolean(settings.assetSlots?.[slot])
            const previewUrl = hasImage ? getThemeEffectsBackgroundUrl(effectiveThemeId, slot, settings.updatedAt) : undefined

            return (
              <div key={slot} className="rounded-lg border p-3 space-y-3">
                <div className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <h5 className="text-sm font-medium">{getThemeEffectsAssetSlotLabel(slot)}</h5>
                    {hasImage && <Badge variant="secondary">Uploaded</Badge>}
                  </div>
                </div>

                <div className="aspect-video overflow-hidden rounded-md border bg-muted/30">
                  {previewUrl ? (
                    <img src={previewUrl} alt="" className="h-full w-full object-cover" />
                  ) : (
                    <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
                      <Image className="mr-2 h-4 w-4" />
                      No image uploaded
                    </div>
                  )}
                </div>

                <div className="flex gap-2">
                  <Button type="button" variant="outline" className="relative" disabled={hasPendingMutation}>
                    <UploadCloud className="mr-2 h-4 w-4" />
                    Upload
                    <Input
                      type="file"
                      accept="image/png,image/jpeg,image/webp"
                      className="absolute inset-0 cursor-pointer opacity-0"
                      onChange={(event) => {
                        void handleUpload(slot, event.target.files?.[0])
                        event.currentTarget.value = ""
                      }}
                    />
                  </Button>
                  <Button type="button" variant="ghost" disabled={!hasImage || hasPendingMutation} onClick={() => void handleDelete(slot)}>
                    <Trash2 className="mr-2 h-4 w-4" />
                    Remove
                  </Button>
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
