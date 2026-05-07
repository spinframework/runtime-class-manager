package controller //nolint:testpackage // whitebox test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rcmv1 "github.com/spinframework/runtime-class-manager/api/v1alpha1"
)

func makeNode(arch string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node1"},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				OperatingSystem: "linux",
				Architecture:    arch,
			},
		},
	}
}

func makeShim(platforms []rcmv1.PlatformArtifact, anonHTTP *rcmv1.AnonHTTPSpec) *rcmv1.Shim {
	return &rcmv1.Shim{
		ObjectMeta: metav1.ObjectMeta{Name: "test-shim"},
		Spec: rcmv1.ShimSpec{
			FetchStrategy: rcmv1.FetchStrategy{
				Platforms: platforms,
				AnonHTTP:  anonHTTP,
			},
		},
	}
}

func TestResolveArtifactForNode(t *testing.T) {
	tests := []struct {
		name         string
		shim         *rcmv1.Shim
		node         *corev1.Node
		wantLocation string
		wantSHA256   string
		wantErr      bool
	}{
		{
			name: "matches platform by Go-style arch (amd64)",
			shim: makeShim([]rcmv1.PlatformArtifact{
				{OS: "linux", Arch: "amd64", Location: "https://example.com/shim-amd64.tar.gz", SHA256: "abc123"},
				{OS: "linux", Arch: "arm64", Location: "https://example.com/shim-arm64.tar.gz", SHA256: "def456"},
			}, nil),
			node:         makeNode("amd64"),
			wantLocation: "https://example.com/shim-amd64.tar.gz",
			wantSHA256:   "abc123",
		},
		{
			name: "matches platform by Go-style arch (arm64)",
			shim: makeShim([]rcmv1.PlatformArtifact{
				{OS: "linux", Arch: "amd64", Location: "https://example.com/shim-amd64.tar.gz"},
				{OS: "linux", Arch: "arm64", Location: "https://example.com/shim-arm64.tar.gz"},
			}, nil),
			node:         makeNode("arm64"),
			wantLocation: "https://example.com/shim-arm64.tar.gz",
		},
		{
			name: "matches platform by uname-style arch (x86_64) against Go-style node (amd64)",
			shim: makeShim([]rcmv1.PlatformArtifact{
				{OS: "linux", Arch: "x86_64", Location: "https://example.com/shim-x86_64.tar.gz"},
			}, nil),
			node:         makeNode("amd64"),
			wantLocation: "https://example.com/shim-x86_64.tar.gz",
		},
		{
			name: "matches platform by uname-style arch (aarch64) against Go-style node (arm64)",
			shim: makeShim([]rcmv1.PlatformArtifact{
				{OS: "linux", Arch: "aarch64", Location: "https://example.com/shim-aarch64.tar.gz", SHA256: "sha-aarch64"},
			}, nil),
			node:         makeNode("arm64"),
			wantLocation: "https://example.com/shim-aarch64.tar.gz",
			wantSHA256:   "sha-aarch64",
		},
		{
			name: "no matching platform returns error",
			shim: makeShim([]rcmv1.PlatformArtifact{
				{OS: "linux", Arch: "arm64", Location: "https://example.com/shim-arm64.tar.gz"},
			}, nil),
			node:    makeNode("amd64"),
			wantErr: true,
		},
		{
			name: "falls back to anonHttp when no platforms specified",
			shim: makeShim(nil, &rcmv1.AnonHTTPSpec{
				Location: "https://example.com/shim-legacy.tar.gz",
			}),
			node:         makeNode("amd64"),
			wantLocation: "https://example.com/shim-legacy.tar.gz",
		},
		{
			name: "anonHttp fallback has empty sha256",
			shim: makeShim(nil, &rcmv1.AnonHTTPSpec{
				Location: "https://example.com/shim-legacy.tar.gz",
			}),
			node:         makeNode("arm64"),
			wantLocation: "https://example.com/shim-legacy.tar.gz",
			wantSHA256:   "",
		},
		{
			name: "platforms take precedence over anonHttp",
			shim: makeShim(
				[]rcmv1.PlatformArtifact{
					{OS: "linux", Arch: "amd64", Location: "https://example.com/shim-platform.tar.gz", SHA256: "plat-sha"},
				},
				&rcmv1.AnonHTTPSpec{Location: "https://example.com/shim-anon.tar.gz"},
			),
			node:         makeNode("amd64"),
			wantLocation: "https://example.com/shim-platform.tar.gz",
			wantSHA256:   "plat-sha",
		},
		{
			name: "platforms specified but no match does NOT fall back to anonHttp",
			shim: makeShim(
				[]rcmv1.PlatformArtifact{
					{OS: "linux", Arch: "arm64", Location: "https://example.com/shim-arm64.tar.gz"},
				},
				&rcmv1.AnonHTTPSpec{Location: "https://example.com/shim-anon.tar.gz"},
			),
			node:    makeNode("amd64"),
			wantErr: true,
		},
		{
			name:    "no fetch source configured returns error",
			shim:    makeShim(nil, nil),
			node:    makeNode("amd64"),
			wantErr: true,
		},
		{
			name: "OS matching is case-insensitive",
			shim: makeShim([]rcmv1.PlatformArtifact{
				{OS: "Linux", Arch: "amd64", Location: "https://example.com/shim.tar.gz"},
			}, nil),
			node:         makeNode("amd64"),
			wantLocation: "https://example.com/shim.tar.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveArtifactForNode(tt.shim, tt.node)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.location != tt.wantLocation {
				t.Errorf("location = %q, want %q", got.location, tt.wantLocation)
			}
			if got.sha256 != tt.wantSHA256 {
				t.Errorf("sha256 = %q, want %q", got.sha256, tt.wantSHA256)
			}
		})
	}
}

func TestMatchesPlatform(t *testing.T) {
	tests := []struct {
		name      string
		platform  rcmv1.PlatformArtifact
		nodeOS    string
		nodeArch  string
		wantMatch bool
	}{
		{
			name:      "exact match linux/amd64",
			platform:  rcmv1.PlatformArtifact{OS: "linux", Arch: "amd64"},
			nodeOS:    "linux",
			nodeArch:  "amd64",
			wantMatch: true,
		},
		{
			name:      "exact match linux/arm64",
			platform:  rcmv1.PlatformArtifact{OS: "linux", Arch: "arm64"},
			nodeOS:    "linux",
			nodeArch:  "arm64",
			wantMatch: true,
		},
		{
			name:      "uname x86_64 matches Go amd64 node",
			platform:  rcmv1.PlatformArtifact{OS: "linux", Arch: "x86_64"},
			nodeOS:    "linux",
			nodeArch:  "amd64",
			wantMatch: true,
		},
		{
			name:      "uname aarch64 matches Go arm64 node",
			platform:  rcmv1.PlatformArtifact{OS: "linux", Arch: "aarch64"},
			nodeOS:    "linux",
			nodeArch:  "arm64",
			wantMatch: true,
		},
		{
			name:      "OS mismatch",
			platform:  rcmv1.PlatformArtifact{OS: "windows", Arch: "amd64"},
			nodeOS:    "linux",
			nodeArch:  "amd64",
			wantMatch: false,
		},
		{
			name:      "arch mismatch",
			platform:  rcmv1.PlatformArtifact{OS: "linux", Arch: "arm64"},
			nodeOS:    "linux",
			nodeArch:  "amd64",
			wantMatch: false,
		},
		{
			name:      "case-insensitive OS",
			platform:  rcmv1.PlatformArtifact{OS: "Linux", Arch: "amd64"},
			nodeOS:    "linux",
			nodeArch:  "amd64",
			wantMatch: true,
		},
		{
			name:      "case-insensitive arch",
			platform:  rcmv1.PlatformArtifact{OS: "linux", Arch: "AMD64"},
			nodeOS:    "linux",
			nodeArch:  "amd64",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesPlatform(tt.platform, tt.nodeOS, tt.nodeArch)
			if got != tt.wantMatch {
				t.Errorf("matchesPlatform() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestNormalizeArch(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"amd64", "x86_64"},
		{"arm64", "aarch64"},
		{"arm", "armv7l"},
		{"ppc64le", "ppc64le"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeArch(tt.input)
			if got != tt.want {
				t.Errorf("normalizeArch(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTrimK8sName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			"k8s-name",
			"k8s-name",
		},
		{
			"k8s-name!",
			"k8s-name",
		},
		{
			"verrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrylong-k8s-nameeeeeeeeeeeeeeeee",
			"verrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrylong-k8s-nameeeeeeeeeeee",
		},
		{
			"verrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrylong-k8s-name----------------",
			"verrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrylong-k8s-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := trimK8sName(tt.input)
			if got != tt.want {
				t.Errorf("trimK8sName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
