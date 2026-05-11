package matching

import "testing"

func TestNormalizeRussianText(t *testing.T) {
	got := Normalize("Ночной Дозор: Ёжик — тест!")
	want := "ночной дозор ежик тест"
	if got != want {
		t.Fatalf("Normalize() = %q, want %q", got, want)
	}
}

func TestSamePersonAcceptsReversedName(t *testing.T) {
	if !SamePerson("Сергей Лукьяненко", "Лукьяненко Сергей") {
		t.Fatal("SamePerson() = false, want true")
	}
}

func TestSamePersonRejectsEmptyValues(t *testing.T) {
	if SamePerson("", "Лукьяненко Сергей") {
		t.Fatal("SamePerson() = true, want false")
	}
}
