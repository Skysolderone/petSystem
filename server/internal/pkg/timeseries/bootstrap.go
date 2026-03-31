package timeseries

import "gorm.io/gorm"

func Bootstrap(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS device_data_points (
			id UUID NOT NULL,
			time TIMESTAMPTZ NOT NULL,
			device_id UUID NOT NULL,
			metric VARCHAR(50) NOT NULL,
			value DOUBLE PRECISION NOT NULL,
			unit VARCHAR(20) NOT NULL DEFAULT '',
			meta JSONB NOT NULL DEFAULT '{}'::jsonb,
			PRIMARY KEY (id, time)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_device_data_device_time ON device_data_points (device_id, time DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_device_data_metric_time ON device_data_points (metric, time DESC)`,
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM pg_available_extensions
				WHERE name = 'timescaledb'
			) THEN
				CREATE EXTENSION IF NOT EXISTS timescaledb;
			END IF;

			IF EXISTS (
				SELECT 1
				FROM pg_proc
				WHERE proname = 'create_hypertable'
			) THEN
				PERFORM create_hypertable(
					'device_data_points',
					'time',
					if_not_exists => TRUE,
					migrate_data => TRUE
				);
			END IF;
		END $$`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}

	return nil
}
