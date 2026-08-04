CREATE TABLE widgets (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'ready')),
  created_at TEXT NOT NULL
);

CREATE TABLE widget_events (
  id INTEGER PRIMARY KEY,
  widget_id INTEGER NOT NULL REFERENCES widgets(id) ON DELETE CASCADE,
  event TEXT NOT NULL
);
