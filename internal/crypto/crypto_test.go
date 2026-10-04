package crypto

import (
	"os"
	"testing"

	"github.com/tsaQB/engsel-go/internal/config"
)

func TestCryptoAlgorithms(t *testing.T) {
	// 1. IV derivation test
	iv := DeriveIV(1700000000000)
	if string(iv) != "bc99b4bf100ad96c" {
		t.Fatalf("IV mismatch: got %s, want bc99b4bf100ad96c", string(iv))
	}

	// 2. PKCS7 Padding & AES CBC Roundtrip
	key := []byte("0123456789abcdef")
	plain := []byte("Hello, World! Testing AES-CBC.")
	ct, err := EncryptAESCBC(key, iv, plain)
	if err != nil {
		t.Fatalf("EncryptAESCBC error: %v", err)
	}

	decrypted, err := DecryptAESCBC(key, iv, ct)
	if err != nil {
		t.Fatalf("DecryptAESCBC error: %v", err)
	}

	if string(decrypted) != string(plain) {
		t.Fatalf("Roundtrip mismatch: got %s, want %s", string(decrypted), string(plain))
	}
}

func TestCryptoParity(t *testing.T) {
	_ = config.LoadEnv("../../.env")
	config.InitConfig()

	if config.Cfg.XDataKey == "" || config.Cfg.XAPIBaseSecret == "" {
		t.Skip("Skipping parity test: .env keys not available (running in CI without credentials)")
	}

	// 1. Encrypt & Decrypt XData test
	pt := `{"test":123}`
	enc, err := EncryptXData(pt, 1700000000000)
	if err != nil {
		t.Fatalf("EncryptXData failed: %v", err)
	}
	if enc != "GUJsBb7oGX_Ecmih-fM6xw==" {
		t.Fatalf("EncryptXData mismatch: got %s, want GUJsBb7oGX_Ecmih-fM6xw==", enc)
	}

	dec, err := DecryptXData(enc, 1700000000000)
	if err != nil {
		t.Fatalf("DecryptXData failed: %v", err)
	}
	if dec != pt {
		t.Fatalf("DecryptXData mismatch: got %s, want %s", dec, pt)
	}

	// 2. MakeXSignature test
	sig := MakeXSignature("token123", "POST", "api/v8/test", 1700000000)
	wantSig := "096bfff78e9ec5a47f9f3e55e1f72a98a4dadcebafb7fbc959245eccf9993b7461a07f4ee38ba9c0f2fe6d54e48de016ff0b911cdd6920a705b70fe8d01c034c"
	if sig != wantSig {
		t.Fatalf("MakeXSignature mismatch: got %s, want %s", sig, wantSig)
	}

	// 3. MakeAxAPISignature test
	axSig := MakeAxAPISignature("2026-10-04T00:00:00.000+0700", "62812345678", "123456", "SMS")
	wantAxSig := "MK9vxyrVaVG74G4XfpgfDtCNkeEcpj01EXOIR2mpbMg="
	if axSig != wantAxSig {
		t.Fatalf("MakeAxAPISignature mismatch: got %s, want %s", axSig, wantAxSig)
	}

	// 4. Circle MSISDN test
	encMs := EncryptCircleMSISDN("628123456789")
	decMs := DecryptCircleMSISDN(encMs)
	if decMs != "628123456789" {
		t.Fatalf("Circle MSISDN roundtrip mismatch: got %s, want 628123456789", decMs)
	}

	// 5. BuildEncryptedField test
	encField := BuildEncryptedField("0123456789abcdef", true)
	if encField != "pxfMBJfNkpT0UTuvczFFCQ==0123456789abcdef" {
		t.Fatalf("BuildEncryptedField mismatch: got %s, want pxfMBJfNkpT0UTuvczFFCQ==0123456789abcdef", encField)
	}

	// 6. AxFingerprint test
	dev := DeviceInfo{
		Manufacturer:   "samsung1234",
		Model:          "SM-N931234",
		Lang:           "en",
		Resolution:     "720x1540",
		TzShort:        "GMT07:00",
		IP:             "192.169.69.69",
		FontScale:      1.0,
		AndroidRelease: "13",
		MSISDN:         "6281398370564",
	}
	fp, err := AxFingerprint(dev, os.Getenv("AX_FP_KEY"))
	if err != nil {
		t.Fatalf("AxFingerprint failed: %v", err)
	}
	wantFP := "IvLRBdVkvJqF5ASd213Y6RdrhrlyJGAqOvT/UU5npEPAVyySfTJlmj92YaclGXz5msAAAqSQkH812hca3QX/aLk05VVJQOsWWpEGyeJ/PTPV5DgCauAInNnhEE5aCdcU"
	if fp != wantFP {
		t.Fatalf("AxFingerprint mismatch: got %s, want %s", fp, wantFP)
	}

	// 7. Payment signature test
	sigPay := MakeXSignaturePayment("acc_tok", 1700000000, "PKG123", "tok_pay", "BALANCE", "BUY_PACKAGE", "payments/api/v8/test")
	wantSigPay := "1721eb3bd48cdc85cd33c86c33b27c852a552a9297a65e7914638918e29f40882753cc936be8f81d331fc0372d0aaf483a1ebe34d887a1b9da0587c824804013"
	if sigPay != wantSigPay {
		t.Fatalf("MakeXSignaturePayment mismatch: got %s, want %s", sigPay, wantSigPay)
	}

	// 8. Loyalty signature test
	sigLoy := MakeXSignatureLoyalty(1700000000, "PKG123", "tok_conf", "gamification/test")
	wantLoy := "5fed66f6572cf426fbe444b62cd6817de8235c2920528e73ae0a155328d8f42d804c54b9d7994874b6a61dfed071ab6bd7be9bf4ded8171d85451591a0e0802c"
	if sigLoy != wantLoy {
		t.Fatalf("MakeXSignatureLoyalty mismatch: got %s, want %s", sigLoy, wantLoy)
	}
}
