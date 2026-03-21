package converter

import (
	"testing"
)

func TestValidateBinaryPath(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		expectedBinaries []string
		wantErr          bool
	}{
		{
			name:             "valid node path",
			path:             "node",
			expectedBinaries: []string{"node"},
			wantErr:          false,
		},
		{
			name:             "valid npx path",
			path:             "npx",
			expectedBinaries: []string{"npx"},
			wantErr:          false,
		},
		{
			name:             "full path to node",
			path:             "/usr/bin/node",
			expectedBinaries: []string{"node"},
			wantErr:          false,
		},
		{
			name:             "full path to npx",
			path:             "/usr/local/bin/npx",
			expectedBinaries: []string{"npx"},
			wantErr:          false,
		},
		{
			name:             "invalid binary",
			path:             "/usr/bin/python",
			expectedBinaries: []string{"node"},
			wantErr:          true,
		},
		{
			name:             "empty path",
			path:             "",
			expectedBinaries: []string{"node"},
			wantErr:          true,
		},
		{
			name:             "multiple allowed binaries",
			path:             "npx",
			expectedBinaries: []string{"node", "npx"},
			wantErr:          false,
		},
		// Note: Windows-style paths work correctly on Windows,
		// but filepath.Base behaves differently cross-platform
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBinaryPath(tc.path, tc.expectedBinaries...)
			if (err != nil) != tc.wantErr {
				t.Errorf("validateBinaryPath(%q, %v) error = %v, wantErr %v",
					tc.path, tc.expectedBinaries, err, tc.wantErr)
			}
		})
	}
}

func TestDefaultZodConvertOptions(t *testing.T) {
	opts := DefaultZodConvertOptions()

	if opts.RefStrategy != "none" {
		t.Errorf("expected RefStrategy 'none', got %q", opts.RefStrategy)
	}
	if opts.NodePath != "node" {
		t.Errorf("expected NodePath 'node', got %q", opts.NodePath)
	}
	if opts.NpxPath != "npx" {
		t.Errorf("expected NpxPath 'npx', got %q", opts.NpxPath)
	}
}
