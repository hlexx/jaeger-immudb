package immudb_storage

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestImmuDbDriver_OpenSession(t *testing.T) {
	type args struct {
		database []string
	}
	tests := []struct {
		name        string
		args        args
		connections int
		wantErr     bool
	}{
		{
			name:        "sessions",
			args:        args{},
			connections: 300,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 1; i < tt.connections; i++ {
				driver, err := New("../../plugin-config.yaml")
				if err != nil {
					t.Error(fmt.Sprintf("[%v] failed to load config file %v", i, err.Error()))
				}
				session, err := driver.OpenSession()
				if err != nil {
					t.Error(fmt.Sprintf("[%v] no session open %v", i, err.Error()))
				}
				if !session.IsConnected() {
					t.Error(fmt.Sprintf("[%v] no connected %v", i, err.Error()))
				}
				err = session.CloseSession(context.Background())
				if err != nil && !tt.wantErr {
					t.Error(fmt.Sprintf("[%v] no session closed %v", i, err.Error()))
				}
			}
		})
	}
}

func TestDriver_Context(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	for {
		select {
		case tm := <-time.After(time.Second):
			t.Logf("time ... %v", tm.Second())
			time.Sleep(time.Second * 3)
		case <-ctx.Done():
			t.Logf("context  done...")
			return
		}
	}
}
