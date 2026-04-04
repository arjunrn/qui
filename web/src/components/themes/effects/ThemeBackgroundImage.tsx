/*
 * Copyright (c) 2025-2026, s0up and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import type { ThemeEffectsBackgroundPosition, ThemeEffectsMobileBackground } from "@/types"

interface ThemeBackgroundImageProps {
  imageUrl?: string
  backgroundPosition: ThemeEffectsBackgroundPosition
  backgroundOpacity: number
  mobileBackground: ThemeEffectsMobileBackground
}

const positionMap: Record<ThemeEffectsBackgroundPosition, string> = {
  top: "50% 12%",
  center: "50% 50%",
  bottom: "50% 88%",
}

export function ThemeBackgroundImage({
  imageUrl,
  backgroundPosition,
  backgroundOpacity,
  mobileBackground,
}: ThemeBackgroundImageProps) {
  if (!imageUrl) {
    return null
  }

  const mobileHiddenClassName = mobileBackground === "disabled" ? "hidden md:block" : ""

  return (
    <div className={`absolute inset-0 ${mobileHiddenClassName}`}>
      <img
        src={imageUrl}
        alt=""
        className="h-full w-full object-cover"
        style={{
          objectPosition: positionMap[backgroundPosition],
          opacity: Math.min(Math.max(backgroundOpacity, 0), 100) / 100,
        }}
      />
    </div>
  )
}
