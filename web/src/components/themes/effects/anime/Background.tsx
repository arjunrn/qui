/*
 * Copyright (c) 2025-2026, s0up and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { animeVisualConfigs } from "@/components/themes/effects/anime/config"
import type { AnimeVariation, ThemeEffectsBackgroundPosition, ThemeEffectsMobileBackground } from "@/types"
import { motion } from "motion/react"

interface AnimeThemeBackgroundProps {
  variation: AnimeVariation
  mode: "light" | "dark"
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

export function AnimeThemeBackground({
  variation,
  mode,
  imageUrl,
  backgroundPosition,
  backgroundOpacity,
  mobileBackground,
}: AnimeThemeBackgroundProps) {
  const palette = animeVisualConfigs[variation].gradient[mode]
  const mobileHiddenClassName = imageUrl && mobileBackground === "disabled" ? "hidden md:block" : ""
  const gradient = `
    radial-gradient(circle at 16% 18%, ${palette[0]} 0%, transparent 40%),
    radial-gradient(circle at 82% 24%, ${palette[1]} 0%, transparent 34%),
    radial-gradient(circle at 50% 82%, ${palette[2]} 0%, transparent 42%),
    linear-gradient(160deg, rgba(255,255,255,0.04), transparent 38%),
    linear-gradient(135deg, rgba(0,0,0,0.12), transparent 55%)
  `

  return (
    <>
      <motion.div
        aria-hidden="true"
        className="absolute inset-0"
        style={{ backgroundImage: gradient }}
        animate={{
          scale: [1, 1.06, 1],
          backgroundPosition: ["0% 0%", "6% 4%", "0% 0%"],
        }}
        transition={{
          duration: 24,
          repeat: Infinity,
          ease: "easeInOut",
        }}
      />
      <motion.div
        aria-hidden="true"
        className="absolute inset-0 opacity-70"
        style={{
          backgroundImage: "linear-gradient(180deg, rgba(255,255,255,0.08), transparent 26%, rgba(0,0,0,0.16) 100%)",
        }}
        animate={{ opacity: [0.58, 0.76, 0.58] }}
        transition={{ duration: 12, repeat: Infinity, ease: "easeInOut" }}
      />
      {imageUrl && (
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
      )}
    </>
  )
}
