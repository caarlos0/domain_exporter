package safeconfig

import (
	"os"
	"reflect"
	"testing"

	"github.com/rs/zerolog/log"
)

func TestNew(t *testing.T) {
	type args struct {
		pathToFile string
	}
	tests := []struct {
		name    string
		args    args
		want    SafeConfig
		wantErr bool
	}{
		{
			name: "Empty file name. Default",
			args: args{
				"",
			},
			want:    SafeConfig{},
			wantErr: false,
		},
		{
			name: "Empty file name",
			args: args{
				"file-which-does-not-exist.yaml",
			},
			want:    SafeConfig{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.pathToFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeConfig_Reload(t *testing.T) {
	tests := []struct {
		name        string
		cfg         SafeConfig
		fileContent string
		wantErr     bool
	}{
		{
			name:        "Load empty file",
			cfg:         SafeConfig{},
			fileContent: "",
			wantErr:     false,
		},
		{
			name:        "yaml is not valid",
			cfg:         SafeConfig{},
			fileContent: "yaml is not correct",
			wantErr:     true,
		},
		{
			name: "Vaidd yaml",
			cfg: SafeConfig{
				Domains: []Domain{{Name: "google.com", Host: ""}},
			},
			fileContent: `
domains:
- google.com`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp(os.TempDir(), "temp.*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			f, err := os.Create(file.Name())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := f.Close(); err != nil {
					t.Fatal(err)
				}
			})
			log.Info().Msg(tt.fileContent)

			_, err = f.WriteString(tt.fileContent)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.Remove(file.Name()); err != nil {
					t.Fatal(err)
				}
			})

			cfg, err := New(file.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeConfig.New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(cfg, tt.cfg) {
				t.Errorf("cfg is not equal:\n got %s\n expected: %s", cfg, tt.cfg)
			}
		})
	}
}
