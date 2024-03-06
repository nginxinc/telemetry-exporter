
package telemetry
/*
This is a generated file. DO NOT EDIT.
*/

import (
	"go.opentelemetry.io/otel/attribute"

	
)

func (d *Data) Attributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue

	

	attrs = append(attrs, attribute.String("ProjectName", d.ProjectName))
	attrs = append(attrs, attribute.String("ProjectVersion", d.ProjectVersion))
	attrs = append(attrs, attribute.String("ProjectArchitecture", d.ProjectArchitecture))
	attrs = append(attrs, attribute.String("ClusterID", d.ClusterID))
	attrs = append(attrs, attribute.String("ClusterVersion", d.ClusterVersion))
	attrs = append(attrs, attribute.String("ClusterPlatform", d.ClusterPlatform))
	attrs = append(attrs, attribute.String("InstallationID", d.InstallationID))
	attrs = append(attrs, attribute.Int64("ClusterNodeCount", d.ClusterNodeCount))
	

	return attrs
}

var _ Exportable = (*Data)(nil)
