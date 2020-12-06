package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

//nolint:tparallel
func TestGetKernelInfo(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name               string
		root               interface{}
		expectedKernelInfo map[string]string
	}{
		{
			name: "windows_services_for_linux",
			root: map[string]interface{}{
				"/proc/sys/kernel": map[string]interface{}{
					"osrelease": "4.19.81-microsoft-standard\n",
					"ostype":    "Linux\n",
					"version":   "#1 SMP Debian 5.2.9-2 (2019-08-21)\n",
				},
			},
			expectedKernelInfo: map[string]string{
				"osrelease": "4.19.81-microsoft-standard",
				"ostype":    "Linux",
				"version":   "#1 SMP Debian 5.2.9-2 (2019-08-21)",
			},
		},
		{
			name: "debian_version_only",
			root: map[string]interface{}{
				"/proc/sys/kernel": map[string]interface{}{
					"version": "#1 SMP Debian 5.2.9-2 (2019-08-21)\n",
				},
			},
			expectedKernelInfo: map[string]string{
				"version": "#1 SMP Debian 5.2.9-2 (2019-08-21)",
			},
		},
		{
			name: "proc_sys_kernel_missing",
			root: map[string]interface{}{
				"/proc/sys": &vfst.Dir{Perm: 0o755},
			},
			expectedKernelInfo: nil,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			t.Cleanup(cleanup)

			actual, err := getKernelInfo(fs)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedKernelInfo, actual)
		})
	}
}

func TestGetOSRelease(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		root     map[string]interface{}
		expected map[string]string
	}{
		{
			name: "fedora",
			root: map[string]interface{}{
				"/etc/os-release": joinLines(
					`NAME=Fedora`,
					`VERSION="17 (Beefy Miracle)"`,
					`ID=fedora`,
					`VERSION_ID=17`,
					`PRETTY_NAME="Fedora 17 (Beefy Miracle)"`,
					`ANSI_COLOR="0;34"`,
					`CPE_NAME="cpe:/o:fedoraproject:fedora:17"`,
					`HOME_URL="https://fedoraproject.org/"`,
					`BUG_REPORT_URL="https://bugzilla.redhat.com/"`,
				),
			},
			expected: map[string]string{
				"NAME":           "Fedora",
				"VERSION":        "17 (Beefy Miracle)",
				"ID":             "fedora",
				"VERSION_ID":     "17",
				"PRETTY_NAME":    "Fedora 17 (Beefy Miracle)",
				"ANSI_COLOR":     "0;34",
				"CPE_NAME":       "cpe:/o:fedoraproject:fedora:17",
				"HOME_URL":       "https://fedoraproject.org/",
				"BUG_REPORT_URL": "https://bugzilla.redhat.com/",
			},
		},
		{
			name: "ubuntu",
			root: map[string]interface{}{
				"/usr/lib/os-release": joinLines(
					`NAME="Ubuntu"`,
					`VERSION="18.04.1 LTS (Bionic Beaver)"`,
					`ID=ubuntu`,
					`ID_LIKE=debian`,
					`PRETTY_NAME="Ubuntu 18.04.1 LTS"`,
					`VERSION_ID="18.04"`,
					`HOME_URL="https://www.ubuntu.com/"`,
					`SUPPORT_URL="https://help.ubuntu.com/"`,
					`BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"`,
					`PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"`,
					`VERSION_CODENAME=bionic`,
					`UBUNTU_CODENAME=bionic`,
				),
			},
			expected: map[string]string{
				"NAME":               "Ubuntu",
				"VERSION":            "18.04.1 LTS (Bionic Beaver)",
				"ID":                 "ubuntu",
				"ID_LIKE":            "debian",
				"PRETTY_NAME":        "Ubuntu 18.04.1 LTS",
				"VERSION_ID":         "18.04",
				"HOME_URL":           "https://www.ubuntu.com/",
				"SUPPORT_URL":        "https://help.ubuntu.com/",
				"BUG_REPORT_URL":     "https://bugs.launchpad.net/ubuntu/",
				"PRIVACY_POLICY_URL": "https://www.ubuntu.com/legal/terms-and-policies/privacy-policy",
				"VERSION_CODENAME":   "bionic",
				"UBUNTU_CODENAME":    "bionic",
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			t.Cleanup(cleanup)

			actual, err := getOSRelease(fs)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

//nolint:paralleltest
func TestParseOSRelease(t *testing.T) {
	for _, tc := range []struct {
		name     string
		s        string
		expected map[string]string
	}{
		{
			name: "fedora",
			s: joinLines(
				`NAME=Fedora`,
				`VERSION="17 (Beefy Miracle)"`,
				`ID=fedora`,
				`VERSION_ID=17`,
				`PRETTY_NAME="Fedora 17 (Beefy Miracle)"`,
				`ANSI_COLOR="0;34"`,
				`CPE_NAME="cpe:/o:fedoraproject:fedora:17"`,
				`HOME_URL="https://fedoraproject.org/"`,
				`BUG_REPORT_URL="https://bugzilla.redhat.com/"`,
			),
			expected: map[string]string{
				"NAME":           "Fedora",
				"VERSION":        "17 (Beefy Miracle)",
				"ID":             "fedora",
				"VERSION_ID":     "17",
				"PRETTY_NAME":    "Fedora 17 (Beefy Miracle)",
				"ANSI_COLOR":     "0;34",
				"CPE_NAME":       "cpe:/o:fedoraproject:fedora:17",
				"HOME_URL":       "https://fedoraproject.org/",
				"BUG_REPORT_URL": "https://bugzilla.redhat.com/",
			},
		},
		{
			name: "ubuntu_with_comments",
			s: joinLines(
				`NAME="Ubuntu"`,
				`VERSION="18.04.1 LTS (Bionic Beaver)"`,
				`ID=ubuntu`,
				`ID_LIKE=debian`,
				`PRETTY_NAME="Ubuntu 18.04.1 LTS"`,
				`VERSION_ID="18.04"`,
				`HOME_URL="https://www.ubuntu.com/"`,
				`SUPPORT_URL="https://help.ubuntu.com/"`,
				`BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"`,
				`PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"`,
				`# comment`,
				``,
				`  # comment`,
				`VERSION_CODENAME=bionic`,
				`UBUNTU_CODENAME=bionic`,
			),
			expected: map[string]string{
				"NAME":               "Ubuntu",
				"VERSION":            "18.04.1 LTS (Bionic Beaver)",
				"ID":                 "ubuntu",
				"ID_LIKE":            "debian",
				"PRETTY_NAME":        "Ubuntu 18.04.1 LTS",
				"VERSION_ID":         "18.04",
				"HOME_URL":           "https://www.ubuntu.com/",
				"SUPPORT_URL":        "https://help.ubuntu.com/",
				"BUG_REPORT_URL":     "https://bugs.launchpad.net/ubuntu/",
				"PRIVACY_POLICY_URL": "https://www.ubuntu.com/legal/terms-and-policies/privacy-policy",
				"VERSION_CODENAME":   "bionic",
				"UBUNTU_CODENAME":    "bionic",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseOSRelease(bytes.NewBufferString(tc.s))
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
