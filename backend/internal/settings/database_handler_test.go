package settings

import "testing"

func TestResolveResetTarget(t *testing.T) {
	t.Run("uses current mode when target is empty", func(t *testing.T) {
		if got, want := resolveResetTarget("", "offline"), "offline"; got != want {
			t.Fatalf("resolveResetTarget() = %q, want %q", got, want)
		}
		if got, want := resolveResetTarget("", "online"), "online"; got != want {
			t.Fatalf("resolveResetTarget() = %q, want %q", got, want)
		}
	})

	t.Run("accepts explicit target override", func(t *testing.T) {
		if got, want := resolveResetTarget("online", "offline"), "online"; got != want {
			t.Fatalf("resolveResetTarget() = %q, want %q", got, want)
		}
		if got, want := resolveResetTarget("offline", "online"), "offline"; got != want {
			t.Fatalf("resolveResetTarget() = %q, want %q", got, want)
		}
	})
}
