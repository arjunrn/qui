/*
 * Copyright (c) 2025-2026, s0up and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

interface ThemeEffectsOverlayProps {
  strength: number
}

export function ThemeEffectsOverlay({ strength }: ThemeEffectsOverlayProps) {
  const normalizedStrength = Math.min(Math.max(strength, 0), 100) / 100
  const baseAlpha = normalizedStrength * 0.44
  const focusAlpha = normalizedStrength * 0.26
  const bottomAlpha = normalizedStrength * 0.34
  const sheenAlpha = 0.025 + (normalizedStrength * 0.03)

  return (
    <div
      aria-hidden="true"
      className="absolute inset-0"
      style={{
        background: `
          linear-gradient(180deg, rgba(7, 10, 18, ${baseAlpha}), rgba(7, 10, 18, ${baseAlpha}) 100%),
          radial-gradient(circle at 54% 44%, rgba(7, 10, 18, ${focusAlpha}) 0%, rgba(7, 10, 18, ${focusAlpha * 0.78}) 34%, transparent 70%),
          linear-gradient(180deg, rgba(255, 255, 255, ${sheenAlpha}), transparent 18%, rgba(7, 10, 18, ${bottomAlpha}) 100%),
          radial-gradient(circle at 50% 8%, rgba(255, 255, 255, ${sheenAlpha * 0.75}) 0%, transparent 24%)
        `,
      }}
    />
  )
}
