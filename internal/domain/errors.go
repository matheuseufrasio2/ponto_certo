package domain

import (
	"errors"
	"strings"
)

// ErrorCode represents a standardized, machine-readable error identifier.
type ErrorCode string

const (
	// 404 - Resource Not Found
	ErrCodeNotFound              ErrorCode = "RESOURCE_NOT_FOUND"
	ErrCodeBarberNotFound        ErrorCode = "BARBER_NOT_FOUND"
	ErrCodeCustomerNotFound      ErrorCode = "CUSTOMER_NOT_FOUND"
	ErrCodeClientNotFound        ErrorCode = "CLIENT_NOT_FOUND"
	ErrCodeServiceNotFound       ErrorCode = "SERVICE_NOT_FOUND"
	ErrCodeBillingNotFound       ErrorCode = "BILLING_NOT_FOUND"
	ErrCodeUserNotFound          ErrorCode = "USER_NOT_FOUND"
	ErrCodeOrganizationNotFound  ErrorCode = "ORGANIZATION_NOT_FOUND"
	ErrCodeUnitNotFound          ErrorCode = "UNIT_NOT_FOUND"
	ErrCodeProductNotFound       ErrorCode = "PRODUCT_NOT_FOUND"
	ErrCodeAppointmentNotFound   ErrorCode = "APPOINTMENT_NOT_FOUND"
	ErrCodeScheduleBlockNotFound ErrorCode = "SCHEDULE_BLOCK_NOT_FOUND"
	ErrCodeEmployeeLeaveNotFound ErrorCode = "EMPLOYEE_LEAVE_NOT_FOUND"

	// 409 - Conflicts & Business Rules
	ErrCodeAppointmentConflict         ErrorCode = "APPOINTMENT_CONFLICT"
	ErrCodeBarberHasFutureAppointments ErrorCode = "BARBER_HAS_FUTURE_APPOINTMENTS"
	ErrCodeBillingAlreadyProcessed     ErrorCode = "BILLING_ALREADY_PROCESSED"
	ErrCodeScheduleBlockConflict       ErrorCode = "SCHEDULE_BLOCK_CONFLICT"
	ErrCodeEmployeeLeaveConflict       ErrorCode = "EMPLOYEE_LEAVE_CONFLICT"
	ErrCodeUnitNameAlreadyExists       ErrorCode = "UNIT_NAME_ALREADY_EXISTS"
	ErrCodeManagerEmailTaken           ErrorCode = "MANAGER_EMAIL_TAKEN"
	ErrCodeServiceNameAlreadyExists    ErrorCode = "SERVICE_NAME_ALREADY_EXISTS"
	ErrCodeProductNameAlreadyExists    ErrorCode = "PRODUCT_NAME_ALREADY_EXISTS"

	// 401 & 403 - Auth, Tenant & Governance
	ErrCodeUnauthorized                ErrorCode = "UNAUTHORIZED"
	ErrCodeInvalidCredentials          ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeInvalidToken                ErrorCode = "INVALID_TOKEN"
	ErrCodeTokenExpired                ErrorCode = "TOKEN_EXPIRED"
	ErrCodeMissingAuthHeader           ErrorCode = "MISSING_AUTHORIZATION_HEADER"
	ErrCodeInvalidAuthHeaderFormat     ErrorCode = "INVALID_AUTHORIZATION_HEADER_FORMAT"
	ErrCodeForbidden                   ErrorCode = "FORBIDDEN"
	ErrCodeCannotImpersonateSelf       ErrorCode = "CANNOT_IMPERSONATE_SELF"
	ErrCodeCannotImpersonateSuperAdmin ErrorCode = "CANNOT_IMPERSONATE_SUPER_ADMIN"
	ErrCodeImpersonationActionBlocked  ErrorCode = "IMPERSONATION_ACTION_BLOCKED"
	ErrCodeTenantInactive              ErrorCode = "TENANT_INACTIVE"
	ErrCodeUnitInactive                ErrorCode = "UNIT_INACTIVE"
	ErrCodeUnitNotInOrg                ErrorCode = "UNIT_NOT_IN_ORG"
	ErrCodeCannotDeactivateSuperAdmin  ErrorCode = "CANNOT_DEACTIVATE_SUPER_ADMIN"
	ErrCodeCannotDeactivateSelf        ErrorCode = "CANNOT_DEACTIVATE_SELF"
	ErrCodeInsufficientRoleHierarchy   ErrorCode = "INSUFFICIENT_ROLE_HIERARCHY"
	ErrCodeBarberNotInOrg              ErrorCode = "BARBER_NOT_IN_ORG"

	// 400 - Validation & Input
	ErrCodeValidationFailed             ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidTimeSlot              ErrorCode = "INVALID_TIME_SLOT"
	ErrCodeInvalidPaymentMethod         ErrorCode = "INVALID_PAYMENT_METHOD"
	ErrCodeInvalidBillingStatus         ErrorCode = "INVALID_BILLING_STATUS"
	ErrCodeInvalidRole                  ErrorCode = "INVALID_ROLE"
	ErrCodeInvalidProductPrice          ErrorCode = "INVALID_PRODUCT_PRICE"
	ErrCodeInvalidProductStock          ErrorCode = "INVALID_PRODUCT_STOCK"
	ErrCodeInvalidAppointmentTransition ErrorCode = "INVALID_APPOINTMENT_TRANSITION"
	ErrCodeInvalidAppointmentTime       ErrorCode = "INVALID_APPOINTMENT_TIME"
	ErrCodeInvalidBarberStatus          ErrorCode = "INVALID_BARBER_STATUS"
	ErrCodeInvalidDateRange             ErrorCode = "INVALID_DATE_RANGE"
	ErrCodeProductNameRequired          ErrorCode = "PRODUCT_NAME_REQUIRED"
	ErrCodeInvalidProductCategory       ErrorCode = "INVALID_PRODUCT_CATEGORY"
	ErrCodeProductCostMinValue          ErrorCode = "PRODUCT_COST_MIN_VALUE"
	ErrCodeUnitIdRequired               ErrorCode = "UNIT_ID_REQUIRED"
	ErrCodeBarberIdRequired             ErrorCode = "BARBER_ID_REQUIRED"
	ErrCodeClientIdRequired             ErrorCode = "CLIENT_ID_REQUIRED"
	ErrCodeServiceIdRequired            ErrorCode = "SERVICE_ID_REQUIRED"
	ErrCodePriceMinValue                ErrorCode = "PRICE_MIN_VALUE"
	ErrCodeInvalidPlan                  ErrorCode = "INVALID_PLAN"
	ErrCodeMonthlyFeeMinValue           ErrorCode = "MONTHLY_FEE_MIN_VALUE"
	ErrCodeOrganizationNameRequired     ErrorCode = "ORGANIZATION_NAME_REQUIRED"
	ErrCodeAdminCredentialsRequired     ErrorCode = "ADMIN_CREDENTIALS_REQUIRED"
	ErrCodeAdminPasswordMinLength       ErrorCode = "ADMIN_PASSWORD_MIN_LENGTH"
	ErrCodeTenantIdRequired             ErrorCode = "TENANT_ID_REQUIRED"
	ErrCodeCannotImpersonateInactive    ErrorCode = "CANNOT_IMPERSONATE_INACTIVE_USER"
	ErrCodeTargetUserIdRequired         ErrorCode = "TARGET_USER_ID_REQUIRED"
	ErrCodeInvalidUUID                  ErrorCode = "INVALID_UUID_FORMAT"
	ErrCodeStartTimeAndEndTimeRequired  ErrorCode = "START_TIME_AND_END_TIME_REQUIRED"
	ErrCodeReasonRequired               ErrorCode = "REASON_REQUIRED"
	ErrCodeUserIdRequired               ErrorCode = "USER_ID_REQUIRED"
	ErrCodeInvalidLeaveType             ErrorCode = "INVALID_LEAVE_TYPE"
	ErrCodeInvalidDateFormat            ErrorCode = "INVALID_DATE_FORMAT"
	ErrCodeInvalidDocument              ErrorCode = "INVALID_DOCUMENT"
	ErrCodeInvalidLogoURL               ErrorCode = "INVALID_LOGO_URL"
	ErrCodeUnitNameRequired             ErrorCode = "UNIT_NAME_REQUIRED"
	ErrCodeInvalidChairCount            ErrorCode = "INVALID_CHAIR_COUNT"
	ErrCodeInvalidCommissionScope       ErrorCode = "INVALID_COMMISSION_SCOPE"

	// 500 - Server
	ErrCodeInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
)

// ValidationError represents structured field-level validation errors.
type ValidationError struct {
	Fields map[string]string `json:"fields"`
}

func (v *ValidationError) Error() string {
	return "validation failed for one or more fields"
}

func NewValidationError() *ValidationError {
	return &ValidationError{Fields: make(map[string]string)}
}

func (v *ValidationError) Add(field, rule string) {
	v.Fields[field] = rule
}

func (v *ValidationError) HasErrors() bool {
	return len(v.Fields) > 0
}

// CodeForError resolves a domain or system error to its canonical ErrorCode.
func CodeForError(err error) ErrorCode {
	if err == nil {
		return ""
	}

	var valErr *ValidationError
	if errors.As(err, &valErr) {
		return ErrCodeValidationFailed
	}

	switch {
	// 404
	case errors.Is(err, ErrNotFound):
		return ErrCodeNotFound
	case errors.Is(err, ErrBarberNotFound):
		return ErrCodeBarberNotFound
	case errors.Is(err, ErrCustomerNotFound):
		return ErrCodeCustomerNotFound
	case errors.Is(err, ErrClientNotFound):
		return ErrCodeClientNotFound
	case errors.Is(err, ErrServiceNotFound):
		return ErrCodeServiceNotFound
	case errors.Is(err, ErrBillingNotFound):
		return ErrCodeBillingNotFound
	case errors.Is(err, ErrUserNotFound):
		return ErrCodeUserNotFound
	case errors.Is(err, ErrOrganizationNotFound):
		return ErrCodeOrganizationNotFound
	case errors.Is(err, ErrUnitNotFound):
		return ErrCodeUnitNotFound
	case errors.Is(err, ErrProductNotFound):
		return ErrCodeProductNotFound
	case errors.Is(err, ErrAppointmentNotFound):
		return ErrCodeAppointmentNotFound
	case errors.Is(err, ErrScheduleBlockNotFound):
		return ErrCodeScheduleBlockNotFound
	case errors.Is(err, ErrEmployeeLeaveNotFound):
		return ErrCodeEmployeeLeaveNotFound

	// 409
	case errors.Is(err, ErrAppointmentConflict):
		return ErrCodeAppointmentConflict
	case errors.Is(err, ErrBarberHasFutureAppointments):
		return ErrCodeBarberHasFutureAppointments
	case errors.Is(err, ErrBillingAlreadyProcessed):
		return ErrCodeBillingAlreadyProcessed
	case errors.Is(err, ErrScheduleBlockConflict):
		return ErrCodeScheduleBlockConflict
	case errors.Is(err, ErrEmployeeLeaveConflict):
		return ErrCodeEmployeeLeaveConflict
	case errors.Is(err, ErrUnitNameAlreadyExists):
		return ErrCodeUnitNameAlreadyExists
	case errors.Is(err, ErrManagerEmailTaken):
		return ErrCodeManagerEmailTaken
	case errors.Is(err, ErrServiceNameAlreadyExists):
		return ErrCodeServiceNameAlreadyExists
	case errors.Is(err, ErrProductNameAlreadyExists):
		return ErrCodeProductNameAlreadyExists

	// 401 & 403
	case errors.Is(err, ErrUnauthorized):
		return ErrCodeUnauthorized
	case errors.Is(err, ErrForbidden):
		return ErrCodeForbidden
	case errors.Is(err, ErrCannotImpersonateSelf):
		return ErrCodeCannotImpersonateSelf
	case errors.Is(err, ErrCannotImpersonateSuperAdmin):
		return ErrCodeCannotImpersonateSuperAdmin
	case errors.Is(err, ErrImpersonationActionBlocked):
		return ErrCodeImpersonationActionBlocked
	case errors.Is(err, ErrTenantInactive):
		return ErrCodeTenantInactive
	case errors.Is(err, ErrUnitInactive):
		return ErrCodeUnitInactive
	case errors.Is(err, ErrUnitNotInOrg):
		return ErrCodeUnitNotInOrg
	case errors.Is(err, ErrCannotDeactivateSuperAdmin):
		return ErrCodeCannotDeactivateSuperAdmin
	case errors.Is(err, ErrCannotDeactivateSelf):
		return ErrCodeCannotDeactivateSelf
	case errors.Is(err, ErrInsufficientRoleHierarchy):
		return ErrCodeInsufficientRoleHierarchy
	case errors.Is(err, ErrBarberNotInOrg):
		return ErrCodeBarberNotInOrg

	// 400
	case errors.Is(err, ErrInvalidTimeSlot):
		return ErrCodeInvalidTimeSlot
	case errors.Is(err, ErrInvalidPaymentMethod):
		return ErrCodeInvalidPaymentMethod
	case errors.Is(err, ErrInvalidBillingStatus):
		return ErrCodeInvalidBillingStatus
	case errors.Is(err, ErrInvalidRole):
		return ErrCodeInvalidRole
	case errors.Is(err, ErrInvalidProductPrice):
		return ErrCodeInvalidProductPrice
	case errors.Is(err, ErrInvalidProductStock):
		return ErrCodeInvalidProductStock
	case errors.Is(err, ErrInvalidAppointmentTransition):
		return ErrCodeInvalidAppointmentTransition
	case errors.Is(err, ErrInvalidAppointmentTime):
		return ErrCodeInvalidAppointmentTime
	case errors.Is(err, ErrInvalidBarberStatus):
		return ErrCodeInvalidBarberStatus
	case errors.Is(err, ErrInvalidDateRange):
		return ErrCodeInvalidDateRange
	case errors.Is(err, ErrInvalidDocument):
		return ErrCodeInvalidDocument
	case errors.Is(err, ErrInvalidLogoURL):
		return ErrCodeInvalidLogoURL
	case errors.Is(err, ErrInvalidUUID):
		return ErrCodeInvalidUUID
	case errors.Is(err, ErrUnitNameRequired):
		return ErrCodeUnitNameRequired
	case errors.Is(err, ErrInvalidChairCount):
		return ErrCodeInvalidChairCount
	case errors.Is(err, ErrInvalidCommissionScope):
		return ErrCodeInvalidCommissionScope
	case errors.Is(err, ErrValidationFailed):
		return ErrCodeValidationFailed

	default:
		// Fallback por string matching
		return CodeForMessage(err.Error())
	}
}

// CodeForMessage resolves common error strings to their respective canonical ErrorCode.
func CodeForMessage(msg string) ErrorCode {
	normalized := strings.TrimSpace(strings.ToLower(msg))

	switch {
	case strings.Contains(normalized, "expected uuid") || strings.Contains(normalized, "format, expected uuid"):
		return ErrCodeInvalidUUID
	case strings.Contains(normalized, "invalid credentials"):
		return ErrCodeInvalidCredentials
	case strings.Contains(normalized, "invalid token"):
		return ErrCodeInvalidToken
	case strings.Contains(normalized, "token expired"):
		return ErrCodeTokenExpired
	case strings.Contains(normalized, "missing authorization header"):
		return ErrCodeMissingAuthHeader
	case strings.Contains(normalized, "invalid authorization header"):
		return ErrCodeInvalidAuthHeaderFormat
	case strings.Contains(normalized, "product name is required"):
		return ErrCodeProductNameRequired
	case strings.Contains(normalized, "invalid product category"):
		return ErrCodeInvalidProductCategory
	case strings.Contains(normalized, "product cost must be greater"):
		return ErrCodeProductCostMinValue
	case strings.Contains(normalized, "unit_id is required") || strings.Contains(normalized, "unit_id parameter is required"):
		return ErrCodeUnitIdRequired
	case strings.Contains(normalized, "barber_id is required") || strings.Contains(normalized, "barber_id path parameter is required"):
		return ErrCodeBarberIdRequired
	case strings.Contains(normalized, "client_id is required"):
		return ErrCodeClientIdRequired
	case strings.Contains(normalized, "service_id is required") || strings.Contains(normalized, "service_id path parameter is required"):
		return ErrCodeServiceIdRequired
	case strings.Contains(normalized, "price_in_cents must be greater"):
		return ErrCodePriceMinValue
	case strings.Contains(normalized, "plan must be"):
		return ErrCodeInvalidPlan
	case strings.Contains(normalized, "monthly fee must be greater"):
		return ErrCodeMonthlyFeeMinValue
	case strings.Contains(normalized, "organization name is required"):
		return ErrCodeOrganizationNameRequired
	case strings.Contains(normalized, "admin email and password are required"):
		return ErrCodeAdminCredentialsRequired
	case strings.Contains(normalized, "password must be at least"):
		return ErrCodeAdminPasswordMinLength
	case strings.Contains(normalized, "tenant id is required"):
		return ErrCodeTenantIdRequired
	case strings.Contains(normalized, "cannot impersonate an inactive user"):
		return ErrCodeCannotImpersonateInactive
	case strings.Contains(normalized, "target user id is required"):
		return ErrCodeTargetUserIdRequired
	case strings.Contains(normalized, "start_time and end_time are required"):
		return ErrCodeStartTimeAndEndTimeRequired
	case strings.Contains(normalized, "reason is required"):
		return ErrCodeReasonRequired
	case strings.Contains(normalized, "user_id is required"):
		return ErrCodeUserIdRequired
	case strings.Contains(normalized, "invalid leave_type"):
		return ErrCodeInvalidLeaveType
	case strings.Contains(normalized, "invalid start_date format") || strings.Contains(normalized, "invalid end_date format"):
		return ErrCodeInvalidDateFormat
	case strings.Contains(normalized, "documento invalido") || strings.Contains(normalized, "cpf ou cnpj"):
		return ErrCodeInvalidDocument
	case strings.Contains(normalized, "url do logotipo invalida"):
		return ErrCodeInvalidLogoURL
	case strings.Contains(normalized, "unit name is required"):
		return ErrCodeUnitNameRequired
	case strings.Contains(normalized, "chair count must be greater"):
		return ErrCodeInvalidChairCount
	case strings.Contains(normalized, "already exists in the organization"):
		return ErrCodeUnitNameAlreadyExists
	case strings.Contains(normalized, "a service with this name already exists"):
		return ErrCodeServiceNameAlreadyExists
	case strings.Contains(normalized, "a product with this name already exists"):
		return ErrCodeProductNameAlreadyExists
	case strings.Contains(normalized, "barbershop unit is inactive"):
		return ErrCodeUnitInactive
	default:
		return ""
	}
}
