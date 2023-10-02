package handlers

import "testing"

func Test_getPathFromRule(t *testing.T) {
	type args struct {
		rule string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "expect /", args: args{rule: "PathPrefix(`/`)"}, want: "/"},
		{name: "expect /tasks", args: args{rule: "PathPrefix(`/tasks`)"}, want: "/tasks"},
		{name: "expect /tasks/one", args: args{rule: "PathPrefix(`/tasks/one`)"}, want: "/tasks/one"},
		{name: "expect /tasks", args: args{rule: "Host(`myhost.com`) && PathPrefix(`/tasks`)"}, want: "/tasks"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPathFromRule(tt.args.rule); got != tt.want {
				t.Errorf("getPathFromRule() = %v, want %v", got, tt.want)
			}
		})
	}
}
