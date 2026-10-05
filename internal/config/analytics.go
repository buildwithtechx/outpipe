package config

type AnalyticsConfig struct {
	TinybirdURL        string `env:"TINYBIRD_URL" json:"tinybirdUrl"`
	TinybirdToken      string `env:"TINYBIRD_TOKEN" json:"-"`
	TinybirdDatasource string `env:"TINYBIRD_DATASOURCE" envDefault:"outpipe_telemetry" json:"tinybirdDatasource"`
	ClickHouseURL      string `env:"CLICKHOUSE_URL" json:"clickHouseUrl"`
	ClickHouseUser     string `env:"CLICKHOUSE_USER" json:"clickHouseUser"`
	ClickHousePassword string `env:"CLICKHOUSE_PASSWORD" json:"-"`
	ClickHouseDatabase string `env:"CLICKHOUSE_DATABASE" envDefault:"default" json:"clickHouseDatabase"`
	ClickHouseTable    string `env:"CLICKHOUSE_TABLE" envDefault:"outpipe_telemetry" json:"clickHouseTable"`
}
