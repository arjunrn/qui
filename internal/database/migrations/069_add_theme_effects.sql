CREATE TABLE IF NOT EXISTS theme_effects_preferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    theme_scope TEXT NOT NULL DEFAULT 'shared',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS theme_effects_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    theme_key TEXT NOT NULL DEFAULT '',
    background_scope TEXT NOT NULL DEFAULT 'shared',
    mode_scope TEXT NOT NULL DEFAULT 'shared',
    particles_mode TEXT NOT NULL DEFAULT 'auto',
    background_position TEXT NOT NULL DEFAULT 'center',
    background_opacity INTEGER NOT NULL DEFAULT 42,
    overlay_strength INTEGER NOT NULL DEFAULT 28,
    mobile_background TEXT NOT NULL DEFAULT 'same',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, theme_key)
);

CREATE TABLE IF NOT EXISTS theme_effects_backgrounds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    theme_key TEXT NOT NULL DEFAULT '',
    slot TEXT NOT NULL,
    path TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, theme_key, slot)
);
