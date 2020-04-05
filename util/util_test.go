package util

import "testing"

func TestCleanUserData(t *testing.T) {
	type args struct {
		word string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "First clean test",
			args: args{
				word: "you",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Second clean test",
			args: args{
				word: "world",
			},
			want:    "world",
			wantErr: false,
		},
		{
			name: "Third clean test",
			args: args{
				word: "Freeze",
			},
			want:    "freez",
			wantErr: false,
		},
		{
			name: "Third clean test",
			args: args{
				word: "subversion",
			},
			want:    "subvers",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CleanUserData(tt.args.word)
			if (err != nil) != tt.wantErr {
				t.Errorf("CleanUserData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CleanUserData() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnglishStopWordChecker(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "First stop checker",
			args: args{
				s: "I'm",
			},
			want: true,
		},
		{
			name: "Second stop checker",
			args: args{
				s: "Hello",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EnglishStopWordChecker(tt.args.s); got != tt.want {
				t.Errorf("EnglishStopWordChecker() = %v, want %v", got, tt.want)
			}
		})
	}
}
