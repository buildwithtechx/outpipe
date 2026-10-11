package config

type AnalyticsConfig struct {
	TinybirdURL        string `env:"TINYBIRD_URL" json:"tinybirdUrl" yaml:"tinybird_url"`
	TinybirdToken      string `env:"TINYBIRD_TOKEN" json:"-" yaml:"tinybird_token"`
	TinybirdDatasource string `env:"TINYBIRD_DATASOURCE" envDefault:"outpipe_telemetry" json:"tinybirdDatasource" yaml:"tinybird_datasource"`
	ClickHouseURL      string `env:"CLICKHOUSE_URL" json:"clickHouseUrl" yaml:"clickhouse_url"`
	ClickHouseUser     string `env:"CLICKHOUSE_USER" json:"clickHouseUser" yaml:"clickhouse_user"`
	ClickHousePassword string `env:"CLICKHOUSE_PASSWORD" json:"-" yaml:"clickhouse_password"`
	ClickHouseDatabase string `env:"CLICKHOUSE_DATABASE" envDefault:"default" json:"clickHouseDatabase" yaml:"clickhouse_database"`
	ClickHouseTable    string `env:"CLICKHOUSE_TABLE" envDefault:"outpipe_telemetry" json:"clickHouseTable" yaml:"clickhouse_table"`
}
