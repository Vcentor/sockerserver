// @Author: Vcentor
// @Date: 2020/10/23 11:16 上午

package processer

import (
	"reflect"
	"testing"
)

func TestJSONProcesser_Unmarshal(t *testing.T) {
	type fields struct {
		RequestIDField string
		ActionField    string
		BodyField      string
		router         map[string]func(string, []byte, interface{})
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Processer
		wantErr bool
	}{
		{
			name: "test-1",
			fields: fields{
				RequestIDField: "requestId",
				ActionField:    "action",
				BodyField:      "body",
				router:         nil,
			},
			args: args{data: []byte(`{
					"requestId": "request_id",
					"action": "GET",
					"body": {"a": "1","b": 2,"c": "3"}
			}`)},
			want: Processer{
				RequestID: "request_id",
				Action:    "GET",
				Body:      []byte(`{"a": "1","b": 2,"c": "3"}`),
			},
			wantErr: false,
		},
		{
			name: "test-2",
			fields: fields{
				RequestIDField: "requestId",
				ActionField:    "action",
				BodyField:      "body",
				router:         nil,
			},
			args: args{data: []byte(`{
					"requestId": "request_id",
					"action": "GET",
					"body": 123
			}`)},
			want: Processer{
				RequestID: "request_id",
				Action:    "GET",
				Body:      []byte(`123`),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSONProcesser{
				RequestIDField: tt.fields.RequestIDField,
				ActionField:    tt.fields.ActionField,
				BodyField:      tt.fields.BodyField,
				router:         tt.fields.router,
			}
			got, err := j.Unmarshal(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONProcesser_Route(t *testing.T) {
	type fields struct {
		RequestIDField string
		ActionField    string
		BodyField      string
		router         map[string]func(string, []byte, interface{})
	}
	type args struct {
		p     Processer
		agent interface{}
	}
	var get = func(requestId string, b []byte, a interface{}) {}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test-route",
			fields: fields{
				RequestIDField: "requestId",
				ActionField:    "action",
				BodyField:      "body",
				router:         map[string]func(string, []byte, interface{}){"get": get},
			},
			args: args{
				p: Processer{
					RequestID: "requestId",
					Action:    "get",
					Body:      []byte("aaa"),
				},
				agent: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSONProcesser{
				RequestIDField: tt.fields.RequestIDField,
				ActionField:    tt.fields.ActionField,
				BodyField:      tt.fields.BodyField,
				router:         tt.fields.router,
			}
			if err := j.Route(tt.args.p, tt.args.agent); (err != nil) != tt.wantErr {
				t.Errorf("Route() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJSONProcesser_RegisterRouter(t *testing.T) {
	type fields struct {
		RequestIDField string
		ActionField    string
		BodyField      string
		router         map[string]func(string, []byte, interface{})
	}
	type args struct {
		action  string
		handler func(string, []byte, interface{})
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test-RegisterRouter",
			fields: fields{
				RequestIDField: "requestId",
				ActionField:    "action",
				BodyField:      "body",
				router:         make(map[string]func(string, []byte, interface{})),
			},
			args: args{
				action:  "get",
				handler: func(a string, b []byte, c interface{}) {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSONProcesser{
				RequestIDField: tt.fields.RequestIDField,
				ActionField:    tt.fields.ActionField,
				BodyField:      tt.fields.BodyField,
				router:         tt.fields.router,
			}
			j.RegisterRouter(tt.args.action, tt.args.handler)
		})
	}
}
