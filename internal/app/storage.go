package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "modernc.org/sqlite"
	"path/filepath"
	"strconv"
	"strings"
)

func openDB(root string) (*sql.DB, *hyperError) {
	db, err := sql.Open("sqlite", filepath.Join(root, hyperDir, "hyper.sqlite"))
	if err != nil {
		return nil, dbError(err)
	}
	return db, nil
}

func ensureSchema(db *sql.DB) *hyperError {
	_, err := db.Exec(`
create table if not exists runs (
  id text primary key,
  project_id text,
  objective text not null,
  stage text not null,
  status text not null,
  started_at text not null,
  ended_at text,
  current_goal_id text,
  summary text
);
create table if not exists goals (
  id text primary key,
  run_id text not null,
  objective text not null,
  scope text,
  non_goals text,
  validation text,
  stop_condition text,
  status text not null,
  created_at text not null,
  completed_at text
);
create table if not exists events (
  id integer primary key autoincrement,
  run_id text,
  goal_id text,
  type text not null,
  payload_json text not null,
  created_at text not null
);
create table if not exists evidence (
  id integer primary key autoincrement,
  run_id text,
  goal_id text,
  kind text,
  path_or_value text,
  summary text,
  status text,
  created_at text not null
);
create table if not exists memories (
  id integer primary key autoincrement,
  project_id text,
  kind text not null,
  text text not null,
  source_event_ids text,
  confidence real,
  quality text,
  created_at text not null,
  last_used_at text,
  stale_at text
);
create table if not exists skill_candidates (
  id integer primary key autoincrement,
  project_id text,
  name text not null,
  summary text,
  trigger_conditions text,
  evidence_count integer default 0,
  success_count integer default 0,
  failure_count integer default 0,
  status text not null,
  created_at text not null,
  promoted_at text
);
create table if not exists harness_candidates (
  id integer primary key autoincrement,
  project_id text,
  name text not null,
  summary text,
  included_patterns text,
  required_validations text,
  status text not null,
  created_at text not null,
  promoted_at text
);`)
	if err != nil {
		return dbError(err)
	}
	if err := ensureColumn(db, "memories", "quality", "text"); err != nil {
		return err
	}
	return nil
}

func ensureColumn(db *sql.DB, table, column, definition string) *hyperError {
	rows, err := db.Query("pragma table_info(" + table + ")")
	if err != nil {
		return dbError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return dbError(err)
		}
		if name == column {
			return nil
		}
	}
	if _, err := db.Exec("alter table " + table + " add column " + column + " " + definition); err != nil {
		return dbError(err)
	}
	return nil
}

func nextID(db *sql.DB, table, prefix string) (string, *hyperError) {
	rows, err := db.Query(fmt.Sprintf("select id from %s where id like ?", table), prefix+"-%")
	if err != nil {
		return "", dbError(err)
	}
	defer rows.Close()
	maxID := 0
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", dbError(err)
		}
		n, _ := strconv.Atoi(strings.TrimPrefix(id, prefix+"-"))
		if n > maxID {
			maxID = n
		}
	}
	return fmt.Sprintf("%s-%04d", prefix, maxID+1), nil
}

func insertRun(db *sql.DB, runID, objective, stage, status, now, currentGoalID, summary string) *hyperError {
	_, err := db.Exec(`insert into runs (id, project_id, objective, stage, status, started_at, current_goal_id, summary) values (?, ?, ?, ?, ?, ?, ?, ?)`, runID, "default", objective, stage, status, now, currentGoalID, summary)
	if err != nil {
		return dbError(err)
	}
	return nil
}

func insertGoal(db *sql.DB, goalID, runID string, ep episode, status, now string) *hyperError {
	_, err := db.Exec(`insert into goals (id, run_id, objective, scope, non_goals, validation, stop_condition, status, created_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?)`, goalID, runID, ep.Objective, ep.Scope, ep.NonGoals, ep.Validation, ep.StopCondition, status, now)
	if err != nil {
		return dbError(err)
	}
	return nil
}

func updateRunAndGoalStatus(db *sql.DB, runID, goalID, status, now string) *hyperError {
	if strings.TrimSpace(goalID) != "" {
		if _, err := db.Exec(`update goals set status = ?, completed_at = ? where id = ?`, status, now, goalID); err != nil {
			return dbError(err)
		}
	}
	if strings.TrimSpace(runID) != "" {
		if _, err := db.Exec(`update runs set status = ?, ended_at = ? where id = ?`, status, now, runID); err != nil {
			return dbError(err)
		}
	}
	return nil
}

func insertEvent(db *sql.DB, event map[string]any) *hyperError {
	eventType, _ := event["type"].(string)
	runID, _ := event["run_id"].(string)
	goalID, _ := event["goal_id"].(string)
	createdAt, _ := event["created_at"].(string)
	if createdAt == "" {
		createdAt = nowISO()
	}
	payload, _ := json.Marshal(event)
	_, err := db.Exec(`insert into events (run_id, goal_id, type, payload_json, created_at) values (?, ?, ?, ?, ?)`, nullableString(runID), nullableString(goalID), eventType, string(payload), createdAt)
	if err != nil {
		return dbError(err)
	}
	return nil
}

func insertMemoryIfNew(db *sql.DB, mem memory) (bool, *hyperError) {
	var count int
	if err := db.QueryRow(`select count(*) from memories where kind = ? and text = ?`, mem.Kind, mem.Text).Scan(&count); err != nil {
		return false, dbError(err)
	}
	if count > 0 {
		return false, nil
	}
	_, err := db.Exec(`insert into memories (project_id, kind, text, source_event_ids, confidence, quality, created_at, last_used_at, stale_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?)`, "default", mem.Kind, mem.Text, nil, mem.Confidence, mem.Quality, nowISO(), nil, nil)
	if err != nil {
		return false, dbError(err)
	}
	return true, nil
}

func countRows(db *sql.DB, table string) (int, *hyperError) {
	var count int
	if err := db.QueryRow("select count(*) from " + table).Scan(&count); err != nil {
		return 0, dbError(err)
	}
	return count, nil
}
