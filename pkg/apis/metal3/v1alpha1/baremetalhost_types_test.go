package v1alpha1

import (
	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHostAvailable(t *testing.T) {
	hostWithError := BareMetalHost{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myhost",
			Namespace: "myns",
		},
	}
	hostWithError.SetErrorMessage("oops something went wrong")

	testCases := []struct {
		Host        BareMetalHost
		Expected    bool
		FailMessage string
	}{
		{
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
			},
			Expected:    true,
			FailMessage: "available host returned not available",
		},
		{
			Host:        hostWithError,
			Expected:    false,
			FailMessage: "host with error returned as available",
		},
		{
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					ConsumerRef: &corev1.ObjectReference{
						Name:      "mymachine",
						Namespace: "myns",
					},
				},
			},
			Expected:    false,
			FailMessage: "host with consumerref returned as available",
		},
		{
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "myhost",
					Namespace:         "myns",
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
			},
			Expected:    false,
			FailMessage: "deleted host returned as available",
		},
	}

	for _, tc := range testCases {
		if tc.Host.Available() != tc.Expected {
			t.Error(tc.FailMessage)
		}
	}
}

func TestHostNeedsHardwareInspection(t *testing.T) {

	testCases := []struct {
		Scenario string
		Host     BareMetalHost
		Expected bool
	}{
		{
			Scenario: "no hardware details",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
			},
			Expected: true,
		},

		{
			Scenario: "host with details",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Status: BareMetalHostStatus{
					HardwareDetails: &HardwareDetails{},
				},
			},
			Expected: false,
		},

		{
			Scenario: "unprovisioned host with consumer",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					ConsumerRef: &corev1.ObjectReference{},
				},
			},
			Expected: true,
		},

		{
			Scenario: "provisioned host",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Status: BareMetalHostStatus{
					Provisioning: ProvisionStatus{
						Image: Image{
							URL: "not-empty",
						},
					},
				},
			},
			Expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Scenario, func(t *testing.T) {
			actual := tc.Host.NeedsHardwareInspection()
			if tc.Expected && !actual {
				t.Error("expected to need hardware inspection")
			}
			if !tc.Expected && actual {
				t.Error("did not expect to need hardware inspection")
			}
		})
	}
}

func TestHostNeedsManualCleaning(t *testing.T) {
	testCases := []struct {
		Scenario   string
		Host       BareMetalHost
		CleanSteps [] nodes.CleanStep
		Expected   bool
	}{

		{
			Scenario: "without cleanSteps in host status and incoming cleanSteps",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
			},
			CleanSteps: []nodes.CleanStep{
				{
					Interface: "fakeInterface",
					Step: "fakeStep",
					Args: map[string]interface{}{
						"fakeKey": "fakeValue",
					},
				},
			},
			Expected: true,
		},

		{
			Scenario: "with cleanSteps in host status and incoming cleanSteps",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Status: BareMetalHostStatus{
					CleanSteps: []nodes.CleanStep{
						{
							Interface: "myInterface",
							Step:      "myStep",
						},
					},
				},
			},
			CleanSteps: []nodes.CleanStep{
				{
					Interface: "fakeInterface",
					Step: "fakeStep",
					Args: map[string]interface{}{
						"fakeKey": "fakeValue",
					},
				},
			},
			Expected: false,
		},

		{
			Scenario: "with cleanSteps in host status and no incoming cleanSteps",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Status: BareMetalHostStatus{
					CleanSteps: []nodes.CleanStep{
						{
							Interface: "myInterface",
							Step:      "myStep",
						},
					},
				},
			},
			CleanSteps: []nodes.CleanStep{},
			Expected: false,
		},

		{
			Scenario: "without cleanSteps in host status and no incoming cleanSteps",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
			},
			CleanSteps: []nodes.CleanStep{},
			Expected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Scenario, func(t *testing.T) {
			actual := tc.Host.NeedsManualCleaning(tc.CleanSteps)
			if tc.Expected && !actual {
				t.Error("expected to need manual cleaning")
			}
			if !tc.Expected && actual {
				t.Error("did not expect to need manual cleaning")
			}
		})
	}
}

func TestHostNeedsProvisioning(t *testing.T) {
	testCases := []struct {
		Scenario string
		Host     BareMetalHost
		Expected bool
	}{

		{
			Scenario: "without image",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					Online: true,
				},
			},
			Expected: false,
		},

		{
			Scenario: "without image url",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					Image:  &Image{},
					Online: true,
				},
			},
			Expected: false,
		},

		{
			Scenario: "with image url, online",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					Image: &Image{
						URL: "not-empty",
					},
					Online: true,
				},
			},
			Expected: true,
		},

		{
			Scenario: "with image url, offline",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					Image: &Image{
						URL: "not-empty",
					},
					Online: false,
				},
			},
			Expected: false,
		},

		{
			Scenario: "already provisioned",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					Image: &Image{
						URL: "not-empty",
					},
					Online: true,
				},
				Status: BareMetalHostStatus{
					Provisioning: ProvisionStatus{
						Image: Image{
							URL: "also-not-empty",
						},
					},
				},
			},
			Expected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Scenario, func(t *testing.T) {
			actual := tc.Host.NeedsProvisioning()
			if tc.Expected && !actual {
				t.Error("expected to need provisioning")
			}
			if !tc.Expected && actual {
				t.Error("did not expect to need provisioning")
			}
		})
	}
}

func TestHostNeedsDeprovisioning(t *testing.T) {
	testCases := []struct {
		Scenario string
		Host     BareMetalHost
		Expected bool
	}{
		{
			Scenario: "with image url, unprovisioned",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					Image: &Image{
						URL: "not-empty",
					},
					Online: true,
				},
			},
			Expected: false,
		},

		{
			Scenario: "with image, unprovisioned",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					Image:  &Image{},
					Online: true,
				},
			},
			Expected: false,
		},

		{
			Scenario: "without, unprovisioned",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					Online: true,
				},
			},
			Expected: false,
		},

		{
			Scenario: "with image url, offline",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					Image: &Image{
						URL: "not-empty",
					},
					Online: false,
				},
			},
			Expected: false,
		},

		{
			Scenario: "provisioned",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					Image: &Image{
						URL: "same",
					},
					Online: true,
				},
				Status: BareMetalHostStatus{
					Provisioning: ProvisionStatus{
						Image: Image{
							URL: "same",
						},
					},
				},
			},
			Expected: false,
		},

		{
			Scenario: "removed image",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					Online: true,
				},
				Status: BareMetalHostStatus{
					Provisioning: ProvisionStatus{
						Image: Image{
							URL: "same",
						},
					},
				},
			},
			Expected: true,
		},

		{
			Scenario: "changed image",
			Host: BareMetalHost{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myhost",
					Namespace: "myns",
				},
				Spec: BareMetalHostSpec{
					Image: &Image{
						URL: "not-empty",
					},
					Online: true,
				},
				Status: BareMetalHostStatus{
					Provisioning: ProvisionStatus{
						Image: Image{
							URL: "also-not-empty",
						},
					},
				},
			},
			Expected: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Scenario, func(t *testing.T) {
			actual := tc.Host.NeedsDeprovisioning()
			if tc.Expected && !actual {
				t.Error("expected to need provisioning")
			}
			if !tc.Expected && actual {
				t.Error("did not expect to need provisioning")
			}
		})
	}
}

func TestHostWasExternallyProvisioned(t *testing.T) {

	for _, tc := range []struct {
		Scenario string
		Host     BareMetalHost
		Expected bool
	}{

		{
			Scenario: "set with image",
			Host: BareMetalHost{
				Spec: BareMetalHostSpec{
					ExternallyProvisioned: true,
					Image: &Image{
						URL: "with-image",
					},
				},
			},
			Expected: false,
		},

		{
			Scenario: "set without image",
			Host: BareMetalHost{
				Spec: BareMetalHostSpec{
					ExternallyProvisioned: true,
				},
			},
			Expected: true,
		},

		{
			Scenario: "set with consumer",
			Host: BareMetalHost{
				Spec: BareMetalHostSpec{
					ExternallyProvisioned: true,
					ConsumerRef:           &corev1.ObjectReference{},
				},
			},
			Expected: true,
		},

		{
			Scenario: "not set without image or consumer",
			Host: BareMetalHost{
				Spec: BareMetalHostSpec{},
			},
			Expected: false,
		},
	} {
		t.Run(tc.Scenario, func(t *testing.T) {
			if tc.Expected && !tc.Host.WasExternallyProvisioned() {
				t.Error("expected to find externally provisioned host")
			}
			if !tc.Expected && tc.Host.WasExternallyProvisioned() {
				t.Error("did not expect to find externally provisioned host")
			}
		})
	}
}
