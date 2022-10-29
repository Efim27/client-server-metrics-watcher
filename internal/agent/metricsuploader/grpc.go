package metricsuploader

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"metrics/internal/agent/statsreader"
	pb "metrics/proto"
)

type MetricsUploaderGRPC struct {
	clientConn *grpc.ClientConn
	client     pb.MetricsClient
}

func NewMetricsUploaderGRPC(addr string) (*MetricsUploaderGRPC, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &MetricsUploaderGRPC{
		clientConn: conn,
		client:     pb.NewMetricsClient(conn),
	}, nil
}

func (m *MetricsUploaderGRPC) Upload(metricsDump statsreader.MetricsDump) (err error) {
	updateMetricsRequest := pb.UpdateMetricsRequest{}

	for metricID, metricValue := range metricsDump.MetricsGauge {
		updateMetricsRequest.Metrics = append(updateMetricsRequest.Metrics, &pb.Metric{
			Metric: &pb.Metric_Gauge{
				Gauge: &pb.MetricGauge{
					Id:    metricID,
					Value: float64(metricValue),
				},
			},
		})
	}

	for metricID, metricValue := range metricsDump.MetricsCounter {
		updateMetricsRequest.Metrics = append(updateMetricsRequest.Metrics, &pb.Metric{
			Metric: &pb.Metric_Counter{
				Counter: &pb.MetricCounter{
					Id:    metricID,
					Delta: int64(metricValue),
				},
			},
		})
	}

	_, err = m.client.UpdateMetrics(context.Background(), &updateMetricsRequest)
	if err != nil {
		return
	}

	return
}
