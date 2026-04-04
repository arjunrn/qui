/*
 * Copyright (c) 2025-2026, s0up and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { useIsMobile } from "@/hooks/useMediaQuery"
import type { ThemeEffectsParticlesMode } from "@/types"
import { useReducedMotion } from "motion/react"
import { memo, useMemo, type CSSProperties } from "react"

interface ThemeParticlesProps {
  mode: "light" | "dark"
  particlesMode: ThemeEffectsParticlesMode
  themeId: string
}

type ParticleStyle = CSSProperties & Record<`--${string}`, string>

function particleShape(index: number) {
  return index % 4 === 0 ? "rounded-full border border-white/20" : "rounded-full"
}

export const ThemeParticles = memo(function ThemeParticles({ mode, particlesMode, themeId }: ThemeParticlesProps) {
  const isMobile = useIsMobile()
  const prefersReducedMotion = useReducedMotion()
  const count = isMobile ? 8 : 18
  const colors = useMemo(() => {
    if (mode === "dark") {
      return [
        "color-mix(in oklab, var(--primary) 34%, white 66%)",
        "color-mix(in oklab, var(--accent) 28%, white 72%)",
        "rgba(255, 255, 255, 0.72)",
      ]
    }

    return [
      "color-mix(in oklab, var(--primary) 58%, transparent)",
      "color-mix(in oklab, var(--foreground) 12%, transparent)",
      "rgba(255, 255, 255, 0.48)",
    ]
  }, [mode])

  const particles = useMemo(() => {
    return Array.from({ length: count }, (_, index) => {
      const size = 4 + ((index * 5) % 8)
      const duration = 18 + ((index * 7) % 18)
      const delay = (index * 0.45) % 5
      const left = (index * 13) % 100
      const sway = (index % 2 === 0 ? 1 : -1) * (8 + ((index * 3) % 10))
      const opacity = mode === "dark" ? 0.22 + ((index % 4) * 0.08) : 0.14 + ((index % 4) * 0.06)

      return {
        key: `${themeId}-${mode}-${index}`,
        size,
        duration,
        delay,
        left,
        sway,
        opacity,
        color: colors[index % colors.length],
      }
    })
  }, [colors, count, mode, themeId])

  if (particlesMode === "off" || (particlesMode === "auto" && prefersReducedMotion)) {
    return null
  }

  return (
    <div aria-hidden="true" className="absolute inset-0 overflow-hidden">
      {particles.map((particle, index) => {
        const trackStyle = {
          left: `${particle.left}%`,
          top: `-${particle.size}px`,
          animationName: "theme-particle-rise",
          animationDuration: `${particle.duration}s`,
          animationDelay: `${particle.delay}s`,
        } satisfies CSSProperties

        const driftStyle = {
          "--theme-particle-sway": `${particle.sway}px`,
          animationDuration: `${Math.max(particle.duration * 0.38, 6)}s`,
          animationDelay: `${particle.delay}s`,
        } satisfies ParticleStyle

        const glyphStyle = {
          width: particle.size,
          height: particle.size,
          background: particle.color,
          opacity: particle.opacity,
          boxShadow: `0 0 10px ${particle.color}`,
          "--theme-particle-opacity": `${particle.opacity}`,
          animationDuration: `${particle.duration}s`,
          animationDelay: `${particle.delay}s`,
          animationName: "theme-particle-fade",
        } satisfies ParticleStyle

        return (
          <span
            key={particle.key}
            className="theme-particle-track absolute block will-change-transform"
            style={trackStyle}
          >
            <span
              className="theme-particle-drift block will-change-transform"
              style={driftStyle}
            >
              <span
                className={`theme-particle-glyph block ${particleShape(index)} blur-[0.4px] will-change-transform`}
                style={glyphStyle}
              />
            </span>
          </span>
        )
      })}
    </div>
  )
})
