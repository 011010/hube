-- Generic typed metadata per note. 009_note_properties only added fixed
-- note columns; this is the key/value store. value holds JSON.
CREATE TABLE note_properties (
  note_id TEXT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  key     TEXT NOT NULL,
  type    TEXT NOT NULL CHECK (type IN ('text', 'number', 'date', 'boolean', 'list')),
  value   TEXT NOT NULL,
  PRIMARY KEY (note_id, key)
);

CREATE INDEX idx_note_properties_key ON note_properties(key);
