package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	influxdb2http "github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

var errNotImplemented = fmt.Errorf("method not implemented")

type mockInflux struct {
	pingErr bool
	pingRun bool
	closeCh chan bool
}

// NOTE: This influxdb2.Client interface is too fatty, must add lot's of methods..
func (i *mockInflux) Setup(ctx context.Context, username, password, org, bucket string, retentionPeriodHours int) (*domain.OnboardingResponse, error) {
	return nil, errNotImplemented
}
func (i *mockInflux) SetupWithToken(ctx context.Context, username, password, org, bucket string, retentionPeriodHours int, token string) (*domain.OnboardingResponse, error) {
	return nil, errNotImplemented
}
func (i *mockInflux) Ready(ctx context.Context) (*domain.Ready, error) {
	return nil, errNotImplemented
}
func (i *mockInflux) Health(ctx context.Context) (*domain.HealthCheck, error) {
	return nil, errNotImplemented
}
func (i *mockInflux) Ping(ctx context.Context) (bool, error) {
	switch {
	case i.pingErr:
		return false, errNotImplemented
	case i.pingRun:
		return false, nil
	}
	return true, nil
}
func (i *mockInflux) Close() {
	close(i.closeCh)
}
func (i *mockInflux) Options() *influxdb2.Options {
	return nil
}
func (i *mockInflux) ServerURL() string {
	return errNotImplemented.Error()
}
func (i *mockInflux) HTTPService() influxdb2http.Service {
	return nil
}
func (i *mockInflux) WriteAPI(org, bucket string) api.WriteAPI {
	return nil
}
func (i *mockInflux) WriteAPIBlocking(org, bucket string) api.WriteAPIBlocking {
	return nil
}
func (i *mockInflux) QueryAPI(org string) api.QueryAPI {
	return nil
}
func (i *mockInflux) AuthorizationsAPI() api.AuthorizationsAPI {
	return nil
}
func (i *mockInflux) OrganizationsAPI() api.OrganizationsAPI {
	return nil
}
func (i *mockInflux) UsersAPI() api.UsersAPI {
	return nil
}
func (i *mockInflux) DeleteAPI() api.DeleteAPI {
	return nil
}
func (i *mockInflux) BucketsAPI() api.BucketsAPI {
	return nil
}
func (i *mockInflux) LabelsAPI() api.LabelsAPI {
	return nil
}
func (i *mockInflux) TasksAPI() api.TasksAPI {
	return nil
}
func (i *mockInflux) APIClient() *domain.Client {
	return nil
}

func TestStartup(t *testing.T) {
	sys := newSystem() // Needs access to the context cancel'r func
	ua := newUnitAsset(*initTemplate(), sys)

	// Bad case: too short collection period
	goodPeriod := ua.CollectionPeriod
	ua.CollectionPeriod = 0
	err := ua.startup()
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	ua.CollectionPeriod = goodPeriod

	// Bad case: error while pinging influxdb server
	ua.influx = &mockInflux{pingErr: true}
	err = ua.startup()
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// Bad case: influxdb not running when pinging
	ua.influx = &mockInflux{pingRun: true}
	err = ua.startup()
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// Good case: startup() enters loop and can be shut down again
	c := make(chan bool)
	ua.influx = &mockInflux{closeCh: c}
	go ua.startup()
	sys.cancel()
	// Wait for startup() to quit it's loop and call cleanup(), which in turn
	// should call influx.Close(). If it times out it failed.
	select {
	case <-c:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("Expected startup to quit and call close(), but timed out")
	}
}
