/*
 * Copyright (c) 2025-2026, s0up and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { animeVisualConfigs } from "@/components/themes/effects/anime/config"
import { useIsMobile } from "@/hooks/useMediaQuery"
import type { AnimeVariation, ThemeEffectsParticlesMode } from "@/types"
import { useReducedMotion } from "motion/react"
import { memo, useMemo, type CSSProperties } from "react"

interface AnimeThemeParticlesProps {
  variation: AnimeVariation
  particlesMode: ThemeEffectsParticlesMode
}

type ParticleStyle = CSSProperties & Record<`--${string}`, string>

function particleStyle(shape: "petal" | "bubble" | "star" | "dust") {
  switch (shape) {
    case "petal":
      return "rounded-[60%_40%_70%_30%/40%_60%_40%_60%]"
    case "bubble":
      return "rounded-full border border-white/35"
    case "star":
      return "rotate-45 rounded-[2px]"
    default:
      return "rounded-full"
  }
}

export const AnimeThemeParticles = memo(function AnimeThemeParticles({ variation, particlesMode }: AnimeThemeParticlesProps) {
  const isMobile = useIsMobile()
  const prefersReducedMotion = useReducedMotion()
  const config = animeVisualConfigs[variation].particles
  const count = isMobile ? 12 : 28
  const particles = useMemo(() => {
    return Array.from({ length: count }, (_, index) => {
      const size = config.size[0] + ((index * 7) % (config.size[1] - config.size[0] + 1))
      const duration = config.duration[0] + ((index * 5) % (config.duration[1] - config.duration[0] + 1))
      const delay = (index * 0.6) % 6
      const left = (index * 17) % 100
      const sway = ((index % 2 === 0 ? 1 : -1) * config.drift)
      const color = config.colors[index % config.colors.length]
      const opacity = 0.28 + ((index % 5) * 0.1)
      const startsLow = variation === "ocean"
      const hasSpin = variation === "midnight" || variation === "sakura"

      return {
        key: `${variation}-${index}`,
        size,
        duration,
        delay,
        left,
        sway,
        color,
        opacity,
        startsLow,
        hasSpin,
      }
    })
  }, [config, count, variation])

  if (particlesMode === "off" || (particlesMode === "auto" && prefersReducedMotion)) {
    return null
  }

  return (
    <div aria-hidden="true" className="absolute inset-0 overflow-hidden">
      {particles.map((particle) => {
        const trackStyle = {
          left: `${particle.left}%`,
          top: particle.startsLow ? "100%" : `-${particle.size}px`,
          animationName: particle.startsLow ? "theme-particle-rise-reverse" : "theme-particle-rise",
          animationDuration: `${particle.duration}s`,
          animationDelay: `${particle.delay}s`,
        } satisfies CSSProperties

        const driftStyle = {
          "--theme-particle-sway": `${particle.sway}px`,
          animationDuration: `${Math.max(particle.duration * 0.42, 6)}s`,
          animationDelay: `${particle.delay}s`,
        } satisfies ParticleStyle

        const glyphStyle = {
          width: particle.size,
          height: config.shape === "petal" ? particle.size * 0.62 : particle.size,
          background: particle.color,
          opacity: particle.opacity,
          boxShadow: config.shape === "star" ? `0 0 12px ${particle.color}` : "none",
          "--theme-particle-opacity": `${particle.opacity}`,
          animationDuration: `${particle.duration}s${particle.hasSpin ? `, ${Math.max(particle.duration * 0.7, 10)}s` : ""}`,
          animationDelay: `${particle.delay}s${particle.hasSpin ? `, ${particle.delay}s` : ""}`,
          animationName: particle.hasSpin ? "theme-particle-fade, theme-particle-spin" : "theme-particle-fade",
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
                className={`theme-particle-glyph block ${particleStyle(config.shape)} ${config.blurClassName} will-change-transform`}
                style={glyphStyle}
              />
            </span>
          </span>
        )
      })}
    </div>
  )
})
