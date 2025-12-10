package exporters

import (
	"database/sql"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type BarbicanUsageExporter struct {
	db      *sql.DB
	secrets *prometheus.Desc
}

func NewBarbicanUsageExporter(db *sql.DB) (*BarbicanUsageExporter, error) {
	return &BarbicanUsageExporter{
		db: db,
		secrets: prometheus.NewDesc(
			"openstack_project_secrets",
			"Total number of secrets per OpenStack project",
			[]string{"project_id"}, nil,
		),
	}, nil
}

func (e *BarbicanUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.secrets
}

func (e *BarbicanUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectMetrics(ch)
}

func (e *BarbicanUsageExporter) collectMetrics(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query(`
		SELECT project_id, COUNT(id) as total_secrets
		FROM secrets 
		WHERE deleted_at IS NULL
		GROUP BY project_id
	`)

	if err != nil {
		log.Println("Error querying Barbican database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var totalSecrets float64
		if err := rows.Scan(&projectID, &totalSecrets); err != nil {
			log.Println("Error scanning Barbican row:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.secrets,
			prometheus.GaugeValue,
			totalSecrets,
			projectID,
		)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in Barbican result set:", err)
	}
}
