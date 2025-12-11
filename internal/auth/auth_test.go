package auth

import "testing"


func TestHashPassword(t *testing.T) {
	testPassword := "test_password"
	_, err := HashPassword(testPassword)

	if err != nil {
		t.Errorf("HashPassword(%v) returned an error in TestHashPassword: %v", testPassword, err)
	}

}

func TestCheckPasswordHashWithCorrectPassword(t *testing.T) {
	testPassword := "test_password"
	
	hash, err := HashPassword(testPassword)
	if err != nil {
		t.Errorf("HashPassword(%v) returned an error in TestCheckPasswordHash: %v", testPassword, err)
	}
	
	match, err := CheckPasswordHash(testPassword, hash)
	if err != nil {
		t.Errorf("CheckPasswordHash(%v, %v) returned an error in TestCheckPasswordHash: %v", testPassword, hash,err)
	}

	if !match {
		t.Errorf("CheckPasswordHash(%v, %v) returned %v instead of %v in TestCheckPasswordHash", testPassword, hash, match, true)
	}
}
func TestCheckPasswordHashWithWrongPassword(t *testing.T) {
	testPassword := "test_password"
	wrongPassword := "wrong_password"
	
	hash, err := HashPassword(testPassword)
	if err != nil {
		t.Errorf("HashPassword(%v) returned an error in TestCheckPasswordHash: %v", testPassword, err)
	}
	
	match, err := CheckPasswordHash(wrongPassword, hash)
	if err != nil {
		t.Errorf("CheckPasswordHash(%v, %v) returned an error in TestCheckPasswordHash: %v", wrongPassword, hash,err)
	}

	if match {
		t.Errorf("CheckPasswordHash(%v, %v) returned %v instead of %v in TestCheckPasswordHash", wrongPassword, hash, match, false)
	}
}
