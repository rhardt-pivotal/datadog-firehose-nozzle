package datadogclient

import (
	"encoding/json"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/DataDog/datadog-firehose-nozzle/util"
)

type Formatter struct{}

func (f Formatter) Format(lookup *util.AppDataLookup, prefix string, maxPostBytes uint32, data map[MetricKey]MetricValue) [][]byte {
	if len(data) == 0 {
		return nil
	}

	var result [][]byte
	seriesBytes := formatMetrics(lookup, prefix, data)
	if uint32(len(seriesBytes)) > maxPostBytes && canSplit(data) {
		metricsA, metricsB := splitPoints(data)
		result = append(result, f.Format(lookup, prefix, maxPostBytes, metricsA)...)
		result = append(result, f.Format(lookup, prefix, maxPostBytes, metricsB)...)

		return result
	}

	result = append(result, seriesBytes)
	return result
}

func formatMetrics(lookup *util.AppDataLookup, prefix string, data map[MetricKey]MetricValue) []byte {
	metrics := []Metric{}
	for key, mVal := range data {
		if key.EventType == events.Envelope_HttpStartStop {
			mVal.Tags = decorateTags(util.Stringify(key.AppId), mVal.Tags, lookup)
		}
		metrics = append(metrics, Metric{
			Metric: prefix + key.Name,
			Points: mVal.Points,
			Type:   "gauge",
			Tags:   mVal.Tags,
			Host:   mVal.Host,
		})
	}

	encodedMetric, _ := json.Marshal(Payload{Series: metrics})
	return encodedMetric
}

func decorateTags(appId string, tags []string, lookup *util.AppDataLookup) []string {
	amd := lookup.LookupAppMetadata(appId)
	tags = appendTagIfNotEmpty(tags, "OrgName", amd.OrgName)
	tags = appendTagIfNotEmpty(tags, "SpaceName", amd.SpaceName)
	tags = appendTagIfNotEmpty(tags, "AppName", amd.AppName)
	return tags

}



//func appendNonHSSMetric(prefix string, key MetricKey, mVal MetricValue, metrics []Metric) []Metric {
//	return append(metrics, Metric{
//		Metric: prefix + key.Name,
//		Points: mVal.Points,
//		Type:   "gauge",
//		Tags:   mVal.Tags,
//		Host:   mVal.Host,
//	})
//}

func canSplit(data map[MetricKey]MetricValue) bool {
	for _, v := range data {
		if len(v.Points) > 1 {
			return true
		}
	}

	return false
}

func splitPoints(data map[MetricKey]MetricValue) (a, b map[MetricKey]MetricValue) {
	a = make(map[MetricKey]MetricValue)
	b = make(map[MetricKey]MetricValue)
	for k, v := range data {
		split := len(v.Points) / 2
		if split == 0 {
			a[k] = MetricValue{
				Tags:   v.Tags,
				Points: v.Points,
			}
			continue
		}

		a[k] = MetricValue{
			Tags:   v.Tags,
			Points: v.Points[:split],
		}
		b[k] = MetricValue{
			Tags:   v.Tags,
			Points: v.Points[split:],
		}
	}
	return a, b
}
