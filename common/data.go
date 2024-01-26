package common

//go:generate go run  github.com/nginxinc/telemetry-exporter/generator github.com/nginxinc/telemetry-exporter/common.Data
type Data struct {
	ProductID string
}
