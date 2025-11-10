import db from './connection.js'

export function runMigrations() {
  db.exec(`
    CREATE TABLE IF NOT EXISTS allocation_requests (
      id TEXT PRIMARY KEY,
      total_minutes INTEGER NOT NULL,
      created_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS task_allocations (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      request_id TEXT NOT NULL,
      task_id TEXT NOT NULL,
      ratio REAL NOT NULL,
      allocated_minutes INTEGER NOT NULL,
      min_minutes INTEGER,
      max_minutes INTEGER,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL,
      FOREIGN KEY (request_id) REFERENCES allocation_requests (id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_task_allocations_request_id
      ON task_allocations (request_id);
    CREATE INDEX IF NOT EXISTS idx_task_allocations_task_id
      ON task_allocations (task_id);
  `)
}
