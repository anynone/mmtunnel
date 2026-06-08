package validate

import "testing"

func TestReservedPath(t *testing.T) {
	if !IsReservedPath("/_tunnel/status") {
		t.Fatal("expected /_tunnel/status to be reserved")
	}
	if IsReservedPath("/api") {
		t.Fatal("did not expect /api to be reserved")
	}
}

func TestPathsConflict(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"/api", "/api", true},
		{"/api", "/api/v1", true},
		{"/api/v1", "/api", true},
		{"/api", "/admin", false},
	}
	for _, tc := range cases {
		if got := PathsConflict(tc.a, tc.b); got != tc.want {
			t.Fatalf("PathsConflict(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestSizeParsesMegabytes(t *testing.T) {
	got, err := Size("100mb", 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != 100*1024*1024 {
		t.Fatalf("got %d", got)
	}
}
