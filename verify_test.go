package registry

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"testing"
)

// swapPublicKey replaces the package-level public key for testing and returns a restore function.
func swapPublicKey(t *testing.T, pub ed25519.PublicKey) {
	t.Helper()
	original, originalExtra := omniviewPublicKey, additionalKeys
	omniviewPublicKey, additionalKeys = pub, nil
	t.Cleanup(func() { omniviewPublicKey, additionalKeys = original, originalExtra })
}

func TestVerifyAdditionalKey(t *testing.T) {
	primary, _, _ := ed25519.GenerateKey(rand.Reader)
	extraPub, extraPriv, _ := ed25519.GenerateKey(rand.Reader)
	swapPublicKey(t, primary)

	checksum := "abc123"
	sigB64 := base64.StdEncoding.EncodeToString(ed25519.Sign(extraPriv, []byte(checksum)))

	if err := VerifyArtifactSignature(checksum, sigB64); err != ErrInvalidSignature {
		t.Fatalf("expected ErrInvalidSignature before the key is added, got: %v", err)
	}
	if err := AddPublicKey(hex.EncodeToString(extraPub)); err != nil {
		t.Fatal(err)
	}
	if err := AddPublicKey(hex.EncodeToString(extraPub)); err != nil {
		t.Fatal(err)
	}
	if got := TrustedPublicKeys(); len(got) != 2 || got[1] != hex.EncodeToString(extraPub) {
		t.Fatalf("unexpected trusted keys: %v", got)
	}
	if err := VerifyArtifactSignature(checksum, sigB64); err != nil {
		t.Fatalf("expected signature from additional key to pass: %v", err)
	}

	// Replacing the primary key keeps the additional keys.
	newPrimary, _, _ := ed25519.GenerateKey(rand.Reader)
	if err := SetPublicKey(hex.EncodeToString(newPrimary)); err != nil {
		t.Fatal(err)
	}
	if err := VerifyArtifactSignature(checksum, sigB64); err != nil {
		t.Fatalf("expected additional key to survive SetPublicKey: %v", err)
	}
}

func TestAddPublicKeyRejectsBadInput(t *testing.T) {
	for _, in := range []string{"", "zz", "abcd"} {
		if err := AddPublicKey(in); err == nil {
			t.Errorf("AddPublicKey(%q): expected error", in)
		}
	}
}

func TestVerifyValidSignature(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	swapPublicKey(t, pub)

	checksum := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	sig := ed25519.Sign(priv, []byte(checksum))
	sigB64 := base64.StdEncoding.EncodeToString(sig)

	if err := VerifyArtifactSignature(checksum, sigB64); err != nil {
		t.Fatalf("expected valid signature to pass: %v", err)
	}
}

func TestVerifyEmptySignature(t *testing.T) {
	err := VerifyArtifactSignature("checksum", "")
	if err != ErrUnsignedArtifact {
		t.Fatalf("expected ErrUnsignedArtifact, got: %v", err)
	}
}

func TestVerifyWrongKey(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	otherPub, _, _ := ed25519.GenerateKey(rand.Reader)
	swapPublicKey(t, otherPub)

	checksum := "abc123"
	sig := ed25519.Sign(priv, []byte(checksum))
	sigB64 := base64.StdEncoding.EncodeToString(sig)

	if err := VerifyArtifactSignature(checksum, sigB64); err == nil {
		t.Fatal("expected verification to fail with wrong key")
	}
}

func TestVerifyTamperedChecksum(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	swapPublicKey(t, pub)

	sig := ed25519.Sign(priv, []byte("original"))
	sigB64 := base64.StdEncoding.EncodeToString(sig)

	if err := VerifyArtifactSignature("tampered", sigB64); err == nil {
		t.Fatal("expected verification to fail for tampered checksum")
	}
}

func TestVerifyMalformedBase64(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	swapPublicKey(t, pub)

	err := VerifyArtifactSignature("checksum", "not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for malformed base64")
	}
}

func TestPublicKeyHexDecode(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	h := hex.EncodeToString(pub)
	decoded, err := hex.DecodeString(h)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != ed25519.PublicKeySize {
		t.Fatalf("expected %d bytes, got %d", ed25519.PublicKeySize, len(decoded))
	}
}
