package validate

import (
	"testing"
)

type TestUser struct {
	Name  string `json:"name" validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=150"`
}

func TestStruct_Valid(t *testing.T) {
	user := TestUser{
		Name:  "John",
		Email: "john@example.com",
		Age:   25,
	}

	err := Struct(user)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestStruct_RequiredField(t *testing.T) {
	user := TestUser{
		Email: "john@example.com",
		Age:   25,
	}

	err := Struct(user)
	if err == nil {
		t.Error("expected error for missing required field")
	}

	errMsg := FirstError(err)
	if errMsg != "name is required" {
		t.Errorf("expected 'name is required', got %q", errMsg)
	}
}

func TestStruct_InvalidEmail(t *testing.T) {
	user := TestUser{
		Name:  "John",
		Email: "invalid-email",
		Age:   25,
	}

	err := Struct(user)
	if err == nil {
		t.Error("expected error for invalid email")
	}

	errMsg := FirstError(err)
	if errMsg != "email must be a valid email" {
		t.Errorf("expected email error, got %q", errMsg)
	}
}

func TestStruct_AgeRange(t *testing.T) {
	tests := []struct {
		name    string
		age     int
		wantErr bool
	}{
		{"valid age", 25, false},
		{"zero age", 0, false},
		{"max age", 150, false},
		{"negative age", -1, true},
		{"over max age", 151, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := TestUser{
				Name:  "John",
				Email: "john@example.com",
				Age:   tt.age,
			}

			err := Struct(user)
			if (err != nil) != tt.wantErr {
				t.Errorf("Struct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVar(t *testing.T) {
	tests := []struct {
		name    string
		field   interface{}
		tag     string
		wantErr bool
	}{
		{"valid email", "test@example.com", "email", false},
		{"invalid email", "not-an-email", "email", true},
		{"required present", "hello", "required", false},
		{"required empty", "", "required", true},
		{"min valid", "hello", "min=3", false},
		{"min invalid", "hi", "min=3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Var(tt.field, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Var() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationErrors(t *testing.T) {
	user := TestUser{
		Name:  "",
		Email: "invalid",
		Age:   -1,
	}

	err := Struct(user)
	if err == nil {
		t.Fatal("expected validation errors")
	}

	errs := ValidationErrors(err)
	if len(errs) != 3 {
		t.Errorf("expected 3 errors, got %d", len(errs))
	}

	if _, ok := errs["name"]; !ok {
		t.Error("expected error for 'name' field")
	}
	if _, ok := errs["email"]; !ok {
		t.Error("expected error for 'email' field")
	}
	if _, ok := errs["age"]; !ok {
		t.Error("expected error for 'age' field")
	}
}
