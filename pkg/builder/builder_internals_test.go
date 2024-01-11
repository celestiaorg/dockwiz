package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanGhURL(t *testing.T) {
	testCases := []struct {
		name          string
		inputURL      string
		expectedURL   string
		expectedError bool
	}{
		{
			name:          "Valid URL with Scheme",
			inputURL:      "https://github.com/your-username/your-repo",
			expectedURL:   "github.com/your-username/your-repo",
			expectedError: false,
		},
		{
			name:          "URL without Scheme",
			inputURL:      "github.com/your-username/your-repo",
			expectedURL:   "github.com/your-username/your-repo",
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleanedURL, err := cleanGhURL(tc.inputURL)

			if tc.expectedError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.Equal(t, tc.expectedURL, cleanedURL, "Cleaned URL should match expected")
			}
		})
	}
}
