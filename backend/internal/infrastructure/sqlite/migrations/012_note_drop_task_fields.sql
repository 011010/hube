-- Notes are editorial content, not work items: priority and due_date carried
-- task semantics that never applied to a note. Status, folder and tags remain.
ALTER TABLE notes DROP COLUMN priority;
ALTER TABLE notes DROP COLUMN due_date;
