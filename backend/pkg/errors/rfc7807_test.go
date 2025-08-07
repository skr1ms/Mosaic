package errors

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProblemDetail(t *testing.T) {
	// Arrange
	errorType := ErrorTypeValidation
	title := "Validation Error"
	detail := "Field is required"
	status := 400

	// Act
	problem := NewProblemDetail(errorType, title, detail, status)

	// Assert
	assert.NotNil(t, problem)
	assert.Equal(t, GetBaseURI()+"validation-error", problem.Type)
	assert.Equal(t, title, problem.Title)
	assert.Equal(t, detail, problem.Detail)
	assert.Equal(t, status, problem.Status)
}

func TestProblemDetail_WithExtension(t *testing.T) {
	// Arrange
	problem := NewProblemDetail(ErrorTypeValidation, "Error", "Detail", 400)

	// Act
	problem.WithExtension("field", "email")

	// Assert
	assert.NotNil(t, problem.Extensions)
	assert.Equal(t, "email", problem.Extensions["field"])
}

func TestProblemDetail_JSONSerialization(t *testing.T) {
	// Arrange
	problem := NewProblemDetail(ErrorTypeValidation, "Validation Error", "Field is required", 400)
	problem.WithExtension("field", "email")
	problem.WithInstance("/api/v1/partners")

	// Act
	jsonData, err := json.Marshal(problem)

	// Assert
	assert.NoError(t, err)

	var decoded ProblemDetail
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, problem.Type, decoded.Type)
	assert.Equal(t, problem.Title, decoded.Title)
	assert.Equal(t, problem.Status, decoded.Status)
	assert.Equal(t, problem.Detail, decoded.Detail)
	assert.Equal(t, problem.Instance, decoded.Instance)
	assert.Equal(t, "email", decoded.Extensions["field"])
}

func TestBadRequestError(t *testing.T) {
	// Act
	err := BadRequestError("Invalid request format")

	// Assert
	assert.Equal(t, "Bad Request", err.Title)
	assert.Equal(t, 400, err.Status)
	assert.Equal(t, "Invalid request format", err.Detail)
	assert.Contains(t, err.Type, "bad-request-error")
}

func TestNotFoundError(t *testing.T) {
	// Act
	err := NotFoundError("Partner")

	// Assert
	assert.Equal(t, "Resource Not Found", err.Title)
	assert.Equal(t, 404, err.Status)
	assert.Equal(t, "Partner not found", err.Detail)
	assert.Contains(t, err.Type, "not-found-error")
}

func TestInternalServerError(t *testing.T) {
	// Act
	err := InternalServerError("Database connection failed")

	// Assert
	assert.Equal(t, "Internal Server Error", err.Title)
	assert.Equal(t, 500, err.Status)
	assert.Equal(t, "Database connection failed", err.Detail)
	assert.Contains(t, err.Type, "internal-server-error")
}

func TestValidationError(t *testing.T) {
	// Act
	err := ValidationError("Email is required")

	// Assert
	assert.Equal(t, "Validation Failed", err.Title)
	assert.Equal(t, 400, err.Status)
	assert.Equal(t, "Email is required", err.Detail)
	assert.Contains(t, err.Type, "validation-error")
}

func TestConflictError(t *testing.T) {
	// Act
	err := ConflictError("Partner already exists")

	// Assert
	assert.Equal(t, "Resource Conflict", err.Title)
	assert.Equal(t, 409, err.Status)
	assert.Equal(t, "Partner already exists", err.Detail)
	assert.Contains(t, err.Type, "conflict-error")
}

func TestAuthenticationError(t *testing.T) {
	// Act
	err := AuthenticationError("Invalid credentials")

	// Assert
	assert.Equal(t, "Authentication Required", err.Title)
	assert.Equal(t, 401, err.Status)
	assert.Equal(t, "Invalid credentials", err.Detail)
	assert.Contains(t, err.Type, "authentication-error")
}

func TestAuthorizationError(t *testing.T) {
	// Act
	err := AuthorizationError("Access denied")

	// Assert
	assert.Equal(t, "Insufficient Permissions", err.Title)
	assert.Equal(t, 403, err.Status)
	assert.Equal(t, "Access denied", err.Detail)
	assert.Contains(t, err.Type, "authorization-error")
}

func TestErrorType_Constants(t *testing.T) {
	assert.Equal(t, ErrorType("validation-error"), ErrorTypeValidation)
	assert.Equal(t, ErrorType("authentication-error"), ErrorTypeAuthentication)
	assert.Equal(t, ErrorType("authorization-error"), ErrorTypeAuthorization)
	assert.Equal(t, ErrorType("not-found-error"), ErrorTypeNotFound)
	assert.Equal(t, ErrorType("conflict-error"), ErrorTypeConflict)
	assert.Equal(t, ErrorType("rate-limit-error"), ErrorTypeRateLimit)
	assert.Equal(t, ErrorType("internal-server-error"), ErrorTypeInternal)
	assert.Equal(t, ErrorType("bad-request-error"), ErrorTypeBadRequest)
	assert.Equal(t, ErrorType("payment-error"), ErrorTypePayment)
}

func TestGetBaseURI(t *testing.T) {
	uri := GetBaseURI()
	assert.Equal(t, "https://api.mosaic.com/errors/", uri)
}

func TestValidationErrorWithFields(t *testing.T) {
	// Arrange
	fields := []ValidationFieldError{
		CreateValidationFieldError("email", "required", "", "Email is required"),
		CreateValidationFieldError("age", "min", 15, "Age must be at least 18"),
	}

	// Act
	err := ValidationErrorWithFields(fields)

	// Assert
	assert.Equal(t, "Validation Failed", err.Title)
	assert.Equal(t, 400, err.Status)
	assert.Contains(t, err.Detail, "validation errors occurred")
	assert.NotNil(t, err.Extensions["validation_errors"])

	validationErrors := err.Extensions["validation_errors"].([]ValidationFieldError)
	assert.Len(t, validationErrors, 2)
	assert.Equal(t, "email", validationErrors[0].Field)
	assert.Equal(t, "required", validationErrors[0].Tag)
}

func TestCreateValidationFieldError(t *testing.T) {
	// Act
	fieldError := CreateValidationFieldError("email", "required", "", "Email is required")

	// Assert
	assert.Equal(t, "email", fieldError.Field)
	assert.Equal(t, "required", fieldError.Tag)
	assert.Equal(t, "", fieldError.Value)
	assert.Equal(t, "Email is required", fieldError.Message)
}

func TestRateLimitError(t *testing.T) {
	// Act
	err := RateLimitError("Rate limit exceeded")

	// Assert
	assert.Equal(t, "Rate Limit Exceeded", err.Title)
	assert.Equal(t, 429, err.Status)
	assert.Equal(t, "Rate limit exceeded", err.Detail)
	assert.Contains(t, err.Type, "rate-limit-error")
}

func TestPaymentError(t *testing.T) {
	// Act
	err := PaymentError("Payment required")

	// Assert
	assert.Equal(t, "Payment Error", err.Title)
	assert.Equal(t, 402, err.Status)
	assert.Equal(t, "Payment required", err.Detail)
	assert.Contains(t, err.Type, "payment-error")
}
