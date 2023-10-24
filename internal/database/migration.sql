CREATE TABLE IF NOT EXISTS `users` (
  `id` integer PRIMARY KEY,
  `email` text,
  `password` text,
  `user_type` integer DEFAULT 0 NOT NULL,
  `last_login` integer DEFAULT 0 NOT NULL,
  `created_at` integer DEFAULT 0 NOT NULL
);

CREATE TABLE IF NOT EXISTS `characters` (
  `id` integer PRIMARY KEY,
  `user_id` integer,
  `name` text  UNIQUE,
  `reputation` integer DEFAULT 0 NOT NULL,
  `health` integer DEFAULT 1 NOT NULL,
  `npc_kill` integer DEFAULT 0 NOT NULL,
  `player_kills` integer DEFAULT 0 NOT NULL,
  `cash` integer DEFAULT 0 NOT NULL,
  `bank` integer DEFAULT 0 NOT NULL,
  `is_admin` integer DEFAULT 0 NOT NULL,
  `hometown` text DEFAULT "" NOT NULL,
  `skill_acc` real DEFAULT 0.0 NOT NULL,
  `skill_hide` real DEFAULT 0.0 NOT NULL,
  `skill_search` real DEFAULT 0.0 NOT NULL,
  `skill_track` real DEFAULT 0.0 NOT NULL,
  `skill_snoop` real DEFAULT 0.0 NOT NULL,
  `location_n` integer DEFAULT 1 NOT NULL,
  `location_e` integer DEFAULT 1 NOT NULL,
  `location_city` integer DEFAULT "" NOT NULL,
  `gang_id` integer DEFAULT 0 NOT NULL,
  `created_at` integer DEFAULT 0 NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS `inventory` (
  `user_id` integer PRIMARY KEY,
  `inventory` text,
  `updated_at` integer,
  FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS `gangs` (
  `id` integer PRIMARY KEY,
  `name` text  UNIQUE,
  `tag` text  UNIQUE,
  `leader_id` integer
);

-- atlas schema apply --url "sqlite://./local.db" --to "file://./internal/database/migration.sql" --dev-url "sqlite://file?mode=memory"
-- atlas schema apply --env turso --to file://internal/database/migration.sql --dev-url "sqlite://file?mode=memory"
