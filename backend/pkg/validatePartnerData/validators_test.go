package validatepartnerdata

import (
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func createMockFieldLevel(value string) validator.FieldLevel {
	return &mockFieldLevel{value: reflect.ValueOf(value)}
}

type mockFieldLevel struct {
	value reflect.Value
}

func (m *mockFieldLevel) Top() reflect.Value      { return reflect.Value{} }
func (m *mockFieldLevel) Parent() reflect.Value   { return reflect.Value{} }
func (m *mockFieldLevel) Field() reflect.Value    { return m.value }
func (m *mockFieldLevel) FieldName() string       { return "" }
func (m *mockFieldLevel) StructFieldName() string { return "" }
func (m *mockFieldLevel) Param() string           { return "" }
func (m *mockFieldLevel) GetTag() string          { return "" }
func (m *mockFieldLevel) ExtractType(field reflect.Value) (reflect.Value, reflect.Kind, bool) {
	return reflect.Value{}, reflect.String, false
}
func (m *mockFieldLevel) GetStructFieldOK() (reflect.Value, reflect.Kind, bool) {
	return reflect.Value{}, reflect.String, false
}
func (m *mockFieldLevel) GetStructFieldOKAdvanced(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool) {
	return reflect.Value{}, reflect.String, false
}
func (m *mockFieldLevel) GetStructFieldOK2() (reflect.Value, reflect.Kind, bool, bool) {
	return reflect.Value{}, reflect.String, false, false
}
func (m *mockFieldLevel) GetStructFieldOKAdvanced2(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool, bool) {
	return reflect.Value{}, reflect.String, false, false
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		expected bool
	}{
		{
			name:     "Valid domain",
			domain:   "example.com",
			expected: true,
		},
		{
			name:     "Valid subdomain",
			domain:   "subdomain.example.com",
			expected: true,
		},
		{
			name:     "Invalid domain with protocol",
			domain:   "https://example.com",
			expected: false,
		},
		{
			name:     "Invalid domain with path",
			domain:   "example.com/path",
			expected: false,
		},
		{
			name:     "Empty domain (allowed)",
			domain:   "",
			expected: true,
		},
		{
			name:     "Invalid characters",
			domain:   "example$.com",
			expected: false,
		},
		{
			name:     "Single word domain (invalid)",
			domain:   "localhost",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fl := createMockFieldLevel(tt.domain)
			result := ValidateDomain(fl)
			assert.Equal(t, tt.expected, result, "Expected %v for domain %s", tt.expected, tt.domain)
		})
	}
}

func TestValidateBusinessEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "Valid business email",
			email:    "contact@company.com",
			expected: true,
		},
		{
			name:     "Invalid - Gmail",
			email:    "user@gmail.com",
			expected: false,
		},
		{
			name:     "Invalid - Mail.ru",
			email:    "user@mail.ru",
			expected: false,
		},
		{
			name:     "Invalid email format",
			email:    "invalid-email",
			expected: false,
		},
		{
			name:     "Empty email (allowed)",
			email:    "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fl := createMockFieldLevel(tt.email)
			result := ValidateBusinessEmail(fl)
			assert.Equal(t, tt.expected, result, "Expected %v for email %s", tt.expected, tt.email)
		})
	}
}

func TestValidatePartnerCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name:     "Valid code",
			code:     "1234",
			expected: true,
		},
		{
			name:     "Valid code with leading zeros",
			code:     "0001",
			expected: true,
		},
		{
			name:     "Invalid - reserved code",
			code:     "0000",
			expected: false,
		},
		{
			name:     "Invalid - too short",
			code:     "123",
			expected: false,
		},
		{
			name:     "Invalid - too long",
			code:     "12345",
			expected: false,
		},
		{
			name:     "Invalid - with letters",
			code:     "123a",
			expected: false,
		},
		{
			name:     "Empty code (allowed)",
			code:     "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fl := createMockFieldLevel(tt.code)
			result := ValidatePartnerCode(fl)
			assert.Equal(t, tt.expected, result, "Expected %v for code %s", tt.expected, tt.code)
		})
	}
}

func TestValidateImageSize(t *testing.T) {
	tests := []struct {
		name     string
		size     string
		expected bool
	}{
		{
			name:     "Valid size 21x30",
			size:     "21x30",
			expected: true,
		},
		{
			name:     "Valid size 30x40",
			size:     "30x40",
			expected: true,
		},
		{
			name:     "Valid size 40x40",
			size:     "40x40",
			expected: true,
		},
		{
			name:     "Valid size 40x50",
			size:     "40x50",
			expected: true,
		},
		{
			name:     "Valid size 40x60",
			size:     "40x60",
			expected: true,
		},
		{
			name:     "Valid size 50x70",
			size:     "50x70",
			expected: true,
		},
		{
			name:     "Invalid size",
			size:     "100x100",
			expected: false,
		},
		{
			name:     "Invalid format",
			size:     "40*50",
			expected: false,
		},
		{
			name:     "Empty size (allowed)",
			size:     "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fl := createMockFieldLevel(tt.size)
			result := ValidateImageSize(fl)
			assert.Equal(t, tt.expected, result, "Expected %v for size %s", tt.expected, tt.size)
		})
	}
}

func TestValidateImageStyle(t *testing.T) {
	tests := []struct {
		name     string
		style    string
		expected bool
	}{
		{
			name:     "Valid style - grayscale",
			style:    "grayscale",
			expected: true,
		},
		{
			name:     "Valid style - skin_tones",
			style:    "skin_tones",
			expected: true,
		},
		{
			name:     "Valid style - pop_art",
			style:    "pop_art",
			expected: true,
		},
		{
			name:     "Valid style - max_colors",
			style:    "max_colors",
			expected: true,
		},
		{
			name:     "Invalid style",
			style:    "rainbow",
			expected: false,
		},
		{
			name:     "Empty style (allowed)",
			style:    "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fl := createMockFieldLevel(tt.style)
			result := ValidateImageStyle(fl)
			assert.Equal(t, tt.expected, result, "Expected %v for style %s", tt.expected, tt.style)
		})
	}
}
