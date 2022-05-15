package validator

import (
	"testing"
)

func TestIsEmpty(t *testing.T) {
	got := IsEmpty("")
	res := true
	if got != res {
		t.Errorf("IsEmpty() = %t; want %t got %t", got, res, got)
	}

	got = IsEmpty("Hello World")
	res = false
	if got != res {
		t.Errorf("IsEmpty(Hello World) = %t; want %t got %t", got, res, got)
	}
}

func TestIsValidName(t *testing.T) {
	got := IsValidName("Sam")
	res := true
	if got != res {
		t.Errorf("IsAlphabet(\"Sam\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidName("Henry V")
	res = true
	if got != res {
		t.Errorf("IsAlphabet(\"Henry V\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidName("Henry .Jr")
	res = true
	if got != res {
		t.Errorf("IsAlphabet(\"Henry .Jr\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidName("Henry 5")
	res = false
	if got != res {
		t.Errorf("IsAlphabet(\"Henry 5\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidName("HenrysØn")
	res = false
	if got != res {
		t.Errorf("IsAlphabet(\"HenrysØn\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidName("<scrip>alert(1);<script>")
	res = false
	if got != res {
		t.Errorf("IsValidName(\"<scrip>alert(1);<script>\") = %t; want %t got %t", got, res, got)
	}
}

func TestIsValidUsername(t *testing.T) {
	got := IsValidUsername("lobby23")
	res := true
	if got != res {
		t.Errorf("IsValidUserName(\"lobby23\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidUsername("lobby_23")
	res = true
	if got != res {
		t.Errorf("IsValidUserName(\"lobby_23\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidUsername("lobby.23")
	res = true
	if got != res {
		t.Errorf("IsValidUserName(\"lobby.23\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidUsername("lobby-23")
	res = true
	if got != res {
		t.Errorf("IsValidUserName(\"lobby-23\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidUsername("lØbby_23")
	res = false
	if got != res {
		t.Errorf("IsValidUserName(\"lØbby_23\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidUsername("_obby_23")
	res = false
	if got != res {
		t.Errorf("IsValidUserName(\"_obby_23\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidUsername("lobby_2_")
	res = false
	if got != res {
		t.Errorf("IsValidUserName(\"_obby_2_\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidUsername("lobby@23")
	res = false
	if got != res {
		t.Errorf("IsValidUserName(\"lobby@23\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidUsername("<scrip>alert(1);<script>")
	res = false
	if got != res {
		t.Errorf("IsValidUserName(\"<scrip>alert(1);<script>\") = %t; want %t got %t", got, res, got)
	}
}

func TestIsValidPassword(t *testing.T) {
	got := IsValidPassword("P@ssw0rd")
	res := true
	if got != res {
		t.Errorf("IsValidPassword(\"P@ssw0rd\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidPassword("Passw0rd")
	res = false
	if got != res {
		t.Errorf("IsValidPassword(\"Passw0rd\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidPassword("p@ssw0rd")
	res = false
	if got != res {
		t.Errorf("IsValidPassword(\"p@ssw0rd\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidPassword("P@ssword")
	res = false
	if got != res {
		t.Errorf("IsValidPassword(\"P@ssword\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidPassword("password")
	res = false
	if got != res {
		t.Errorf("IsValidPassword(\"password\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidPassword("password>")
	res = false
	if got != res {
		t.Errorf("IsValidPassword(\"password>\") = %t; want %t got %t", got, res, got)
	}

	got = IsValidName("<scrip>alert(1);<script>")
	res = false
	if got != res {
		t.Errorf("IsValidPassword(\"<scrip>alert(1);<script>\") = %t; want %t got %t", got, res, got)
	}
}

func TestIsMobileNumber(t *testing.T) {
	got := IsMobileNumber("98461564")
	res := true
	if got != res {
		t.Errorf("IsMobileNumber(98461564) = %t; want %t got %t", got, res, got)
	}

	got = IsMobileNumber("86995154")
	res = true
	if got != res {
		t.Errorf("IsMobileNumber(86995154) = %t; want %t got %t", got, res, got)
	}

	got = IsMobileNumber("70472710")
	res = false
	if got != res {
		t.Errorf("IsMobileNumber(70472710) = %t; want %t got %t", got, res, got)
	}

	got = IsMobileNumber("8795509")
	res = false
	if got != res {
		t.Errorf("IsMobileNumber(8795509) = %t; want %t got %t", got, res, got)
	}

	got = IsMobileNumber("<scrip>alert(1);<script>")
	res = false
	if got != res {
		t.Errorf("IsMobileNumber(\"<scrip>alert(1);<script>\") = %t; want %t got %t", got, res, got)
	}
}
