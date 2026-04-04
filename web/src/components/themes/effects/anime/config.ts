/*
 * Copyright (c) 2025-2026, s0up and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import type { AnimeVariation } from "@/types"

type AnimeVisualConfig = {
  gradient: {
    light: string[]
    dark: string[]
  }
  particles: {
    colors: string[]
    shape: "petal" | "bubble" | "star" | "dust"
    drift: number
    duration: [number, number]
    size: [number, number]
    blurClassName: string
  }
}

export const animeVisualConfigs: Record<AnimeVariation, AnimeVisualConfig> = {
  sakura: {
    gradient: {
      light: ["rgba(255, 204, 232, 0.82)", "rgba(241, 180, 255, 0.48)", "rgba(255, 248, 252, 0.72)"],
      dark: ["rgba(82, 17, 54, 0.92)", "rgba(159, 88, 196, 0.42)", "rgba(25, 9, 28, 0.95)"],
    },
    particles: {
      colors: ["#f9bdd9", "#ffdff1", "#f29bc7"],
      shape: "petal",
      drift: 30,
      duration: [16, 26],
      size: [10, 18],
      blurClassName: "blur-[0.2px]",
    },
  },
  ocean: {
    gradient: {
      light: ["rgba(144, 228, 255, 0.68)", "rgba(95, 173, 255, 0.36)", "rgba(240, 252, 255, 0.74)"],
      dark: ["rgba(5, 36, 73, 0.96)", "rgba(0, 122, 163, 0.34)", "rgba(3, 13, 32, 0.96)"],
    },
    particles: {
      colors: ["#a7ecff", "#76d8ff", "#d7fbff"],
      shape: "bubble",
      drift: 18,
      duration: [14, 24],
      size: [8, 16],
      blurClassName: "blur-[0.3px]",
    },
  },
  midnight: {
    gradient: {
      light: ["rgba(193, 182, 255, 0.56)", "rgba(125, 119, 255, 0.28)", "rgba(243, 244, 255, 0.82)"],
      dark: ["rgba(11, 8, 31, 0.98)", "rgba(83, 69, 172, 0.32)", "rgba(2, 2, 11, 0.98)"],
    },
    particles: {
      colors: ["#f4f1ff", "#c6bdff", "#8aa5ff"],
      shape: "star",
      drift: 10,
      duration: [20, 34],
      size: [4, 10],
      blurClassName: "blur-[0.1px]",
    },
  },
  sunset: {
    gradient: {
      light: ["rgba(255, 197, 132, 0.78)", "rgba(255, 119, 146, 0.4)", "rgba(255, 248, 236, 0.7)"],
      dark: ["rgba(64, 18, 7, 0.96)", "rgba(184, 79, 55, 0.3)", "rgba(22, 7, 4, 0.98)"],
    },
    particles: {
      colors: ["#ffd58e", "#ffb46b", "#fff0c9"],
      shape: "dust",
      drift: 16,
      duration: [18, 28],
      size: [5, 12],
      blurClassName: "blur-[0.6px]",
    },
  },
}
