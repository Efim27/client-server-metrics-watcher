package grpc

import (
	"context"

	"github.com/asaskevich/govalidator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"metrics/internal/server/storage"
	pb "metrics/proto"
)

type MetricsService struct {
	storage storage.MetricStorager
	pb.UnimplementedMetricsServer
}

func NewMetricsService(storage storage.MetricStorager) *MetricsService {
	return &MetricsService{
		storage: storage,
	}
}

func (s *MetricsService) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.Empty, error) {
	var MetricBatch []storage.Metric

	if len(in.Metrics) == 0 {
		return nil, status.Errorf(codes.OutOfRange, "empty metric list")
	}

	for _, metric := range in.Metrics {
		switch metricOne := metric.Metric.(type) {
		case *pb.Metric_Gauge:
			MetricBatch = append(MetricBatch, storage.Metric{
				ID: metricOne.Gauge.Id,
				MetricValue: storage.MetricValue{
					MType: storage.MeticTypeGauge,
					Value: &metricOne.Gauge.Value,
				},
			})
		case *pb.Metric_Counter:
			MetricBatch = append(MetricBatch, storage.Metric{
				ID: metricOne.Counter.Id,
				MetricValue: storage.MetricValue{
					MType: storage.MeticTypeCounter,
					Delta: &metricOne.Counter.Delta,
				},
			})
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unknown metric type")
		}
	}

	//Validation
	var err error
	for _, OneMetric := range MetricBatch {
		_, err = govalidator.ValidateStruct(OneMetric)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
	}

	err = s.storage.UpdateManySliceMetric(MetricBatch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.Empty{}, nil
}
