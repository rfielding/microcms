CREATE TABLE `filemeta` (
      `id` INTEGER PRIMARY KEY AUTOINCREMENT,
      `cmd` TEXT,
      `path` TEXT,
      `name` TEXT,
      `contentType` TEXT,
      `contentSize` INTEGER
);

CREATE VIRTUAL TABLE `filesearch` USING FTS5(
      `id`,
      `cmd`,
      `path`,
      `name`,
      `part`,
      `original_path`,
      `original_name`,
      `content`
);
