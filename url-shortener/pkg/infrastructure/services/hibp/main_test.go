package hibp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestHIBP_CheckPasswordIsBreached(t *testing.T) {
	response := []string{
		"00053EB46174159B02663FE71F23794404E:5",
		"00259EE48BC504318E9500D811D1E1E5E2B:66",
		"004181220B092B49B51A26546018E1505D9:1",
		"006BA25BCE7ABE8FEDBA7DF33DEC735BFBF:1",
		"0088ABAAFE4C122548764FBB847F05BBF15",
		"e658aa36f5d42623c41bb17b84a6dd7e51d:1",
	}

	type args struct {
		server *httptest.Server
		path   string
	}

	tests := []struct {
		name     string
		password string
		want     bool
		wantErr  bool
		setup    func() args
	}{
		{
			name:     "happy case: password is in breach",
			password: "collins",
			setup: func() args {
				server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "text/plain")
					w.Write([]byte(strings.Join(response, "\r\n")))
				}))

				return args{server: server, path: "/range"}
			},
			want:    true,
			wantErr: false,
		},
		{
			name:     "happy case: password is not in breach",
			password: "collins",
			setup: func() args {
				server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "text/plain")
					w.Write([]byte(strings.Join([]string{
						"00053EB46174159B02663FE71F23794404E:5",
						"00259EE48BC504318E9500D811D1E1E5E2B:66",
						"004181220B092B49B51A26546018E1505D9:1",
					}, "\r\n")))
				}))

				return args{server: server, path: "/range"}
			},
			want:    false,
			wantErr: false,
		},
		{
			name:     "sad case: failed creating request",
			password: "collins",
			setup: func() args {
				server := httptest.NewUnstartedServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				}))

				return args{server: server, path: ""}
			},
			want:    false,
			wantErr: true,
		},
		{
			name:     "sad case: non-200 status code",
			password: "collins",
			setup: func() args {
				server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				}))

				return args{server: server, path: "/range"}
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.setup()

			args.server.Start()
			defer args.server.Close()

			h := HIBP{
				client:  args.server.Client(),
				baseURL: args.server.URL + args.path,
			}

			got, gotErr := h.CheckPasswordIsBreached(context.Background(), tt.password)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("CheckPasswordIsBreached() error = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("CheckPasswordIsBreached() = %v, want %v", got, tt.want)

				return
			}
		})
	}
}
