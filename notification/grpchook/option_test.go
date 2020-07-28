package grpchook_test

import (
	"context"
	"errors"
	"testing"

	"github.com/SecuritasCrimePrediction/apitools-go/notification/grpchook"
	"github.com/gogo/status"
	"google.golang.org/grpc/codes"
)

func Test_Config_ShouldNotifyForErr(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  grpchook.Config
		in   error
		want bool
	}{
		{"default should not notify on success", grpchook.Config{}, nil, false},
		{"default should notify on error", grpchook.Config{}, status.Error(codes.NotFound, "a"), true},
		{
			"should notify when matching error codes",
			grpchook.Config{
				ErrorCodes: []codes.Code{codes.NotFound},
			},
			status.Error(codes.NotFound, "a"),
			true,
		},
		{
			"should not notify when error codes provided but not matching",
			grpchook.Config{
				ErrorCodes: []codes.Code{codes.InvalidArgument},
			},
			status.Error(codes.NotFound, "a"),
			false,
		},
		{"should notify on success when NotifyOnSuccess=true", grpchook.Config{NotifyOnSuccess: true}, nil, true},
		{"should skip errors when SkipErrors=true", grpchook.Config{SkipErrors: true}, errors.New("a"), false},
		{
			"should use custom func",
			grpchook.Config{
				// These settings should always notify, but custom func always returns false
				NotifyOnSuccess:    true,
				SkipErrors:         false,
				ErrorCodes:         nil,
				CustomShouldNotify: func(ctx context.Context, err error) bool { return false },
			},
			errors.New("a"),
			false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cfg.ShouldNotifyForErr(context.Background(), tc.in)
			if got != tc.want {
				t.Errorf("expected: %t, got: %v", tc.want, got)
			}
		})
	}
}
