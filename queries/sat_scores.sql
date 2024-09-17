-- name: UpdateSATScoreRanks :exec
-- Update the rank for all sat scores
UPDATE sat_scores
SET rank = (
    SELECT COUNT(*) + 1
    FROM sat_scores AS s2
    WHERE s2.sat_score > sat_scores.sat_score
);

-- name: GetSATScores :many
-- Get all sat scores with rank
SELECT * FROM sat_scores ORDER BY rank;

-- name: GetSATScoreByName :one
-- Get a single SAT score record by name
SELECT * FROM sat_scores WHERE name = ?;

-- name: GetNameBySubstring :many
-- Retrieve all names that contain a specific substring
SELECT * FROM sat_scores
WHERE name LIKE ?;

-- name: InsertSATScore :one
-- Add a record to the sat score table
INSERT INTO sat_scores (
    name, address, city, country, pincode, sat_score, passed, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: UpdateSATScore :one
-- Update an existing SAT score record
UPDATE sat_scores
SET sat_score = ?, 
    passed = ?, 
    updated_at = CURRENT_TIMESTAMP
WHERE name = ?
RETURNING *;

-- name: DeleteSATScore :exec
-- Delete an SAT score record by name
DELETE FROM sat_scores WHERE name = ?;