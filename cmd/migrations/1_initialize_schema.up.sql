CREATE TABLE IF NOT EXISTS users (
		uuid TEXT PRIMARY KEY,
		login TEXT,
		hash_pass TEXT,
		UNIQUE (login)
	  );

CREATE TABLE IF NOT EXISTS audiofiles (
		file_id TEXT PRIMARY KEY,
		file_name TEXT,
		user_id TEXT,
		uploaded_at TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(uuid)
	  );

CREATE TABLE IF NOT EXISTS asr (
		uuid TEXT PRIMARY KEY,
		file_id TEXT,
		asr TEXT,
		status TEXT,
		FOREIGN KEY (file_id) REFERENCES audiofiles(file_id)
	  );

CREATE TABLE IF NOT EXISTS result_asr (
		uuid TEXT,
		channel_tag TEXT,
		text TEXT,
		start_time REAL,
		end_time REAL,
		FOREIGN KEY (uuid) REFERENCES asr(uuid)
	  );

CREATE TABLE IF NOT EXISTS quality_control (
		uuid TEXT PRIMARY KEY,
		file_id TEXT,
		channel_tag TEXT,
		text TEXT,
		UNIQUE (file_id, channel_tag),
		FOREIGN KEY (file_id) REFERENCES audiofiles(file_id)
	  );