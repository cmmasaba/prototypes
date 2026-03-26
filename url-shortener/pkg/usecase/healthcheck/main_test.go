package healthcheck

import (
	"context"
	"reflect"
	"testing"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/healthcheck/mocks"
	"github.com/stretchr/testify/mock"
)

func TestUsecaseImplHealth_HealthCheck(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name  string
		want  map[string]bool
		setup func(repo *mocks.Mockrepository, queue *mocks.Mockqueue) args
	}{
		{
			name: "happy case: successful health check",
			want: map[string]bool{
				"database":   true,
				"tasksqueue": true,
			},
			setup: func(repo *mocks.Mockrepository, queue *mocks.Mockqueue) args {
				ctx := context.Background()

				repo.EXPECT().PingDB(mock.Anything).RunAndReturn(func(_ context.Context) bool {
					return true
				})
				queue.EXPECT().PingQueue(mock.Anything).RunAndReturn(func(_ context.Context) bool {
					return true
				})

				return args{ctx: ctx}
			},
		},
		{
			name: "sad case: db health check failed",
			want: map[string]bool{
				"database":   false,
				"tasksqueue": true,
			},
			setup: func(repo *mocks.Mockrepository, queue *mocks.Mockqueue) args {
				ctx := context.Background()

				repo.EXPECT().PingDB(mock.Anything).RunAndReturn(func(_ context.Context) bool {
					return false
				})
				queue.EXPECT().PingQueue(mock.Anything).RunAndReturn(func(_ context.Context) bool {
					return true
				})

				return args{ctx: ctx}
			},
		},
		{
			name: "sad case: task queue health check failed",
			want: map[string]bool{
				"database":   true,
				"tasksqueue": false,
			},
			setup: func(repo *mocks.Mockrepository, queue *mocks.Mockqueue) args {
				ctx := context.Background()

				repo.EXPECT().PingDB(mock.Anything).RunAndReturn(func(_ context.Context) bool {
					return true
				})
				queue.EXPECT().PingQueue(mock.Anything).RunAndReturn(func(_ context.Context) bool {
					return false
				})

				return args{ctx: ctx}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockrepository(t)
			queue := mocks.NewMockqueue(t)
			args := tt.setup(repo, queue)

			u := New(repo, queue)

			got := u.HealthCheck(args.ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HealthCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}
