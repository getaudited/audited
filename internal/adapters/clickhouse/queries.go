package clickhouse

const queryFindByID = `
	SELECT 
		id,
		source_id,
		version,
		action,
		actor_id,
		actor_type,
		actor_name,
		actor_metadata,
		context_location,
		context_user_agent,
		metadata,
		occurred_at,
		targets.internal_id,
		targets.id,
		targets.name,
		targets.type,
		targets.metadata
	FROM events
	WHERE id = ?
`

const querySaveEvent = `
	INSERT INTO events (
		id,
		source_id,
		version,
		action,
		actor_id,
		actor_type,
		actor_name,
		actor_metadata,
		context_location,
		context_user_agent,
		metadata,
		occurred_at,
		targets.internal_id,
		targets.id,
		targets.name,
		targets.type,
		targets.metadata)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
`
