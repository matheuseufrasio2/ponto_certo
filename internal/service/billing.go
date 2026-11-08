package service

import (
	"context"
	"encoding/json"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/repository"
)

// BillingService manages billing and financial transactions.
type BillingService struct {
	billingRepo   repository.BillingRepository
	auditRecorder AuditLogRecorder
}

// NewBillingService instantiates BillingService.
func NewBillingService(billingRepo repository.BillingRepository) *BillingService {
	return &BillingService{billingRepo: billingRepo}
}

// SetAuditRecorder assigns the audit log recorder.
func (s *BillingService) SetAuditRecorder(recorder AuditLogRecorder) {
	s.auditRecorder = recorder
}

// Checkout executes final payment processing for an appointment.
func (s *BillingService) Checkout(ctx context.Context, input domain.CheckoutInput, cashierUserID string) (*domain.Billing, error) {
	if input.AppointmentID == "" {
		return nil, domain.ErrAppointmentNotFound
	}
	if !input.PaymentMethod.IsValid() {
		return nil, domain.ErrInvalidPaymentMethod
	}

	billing, err := s.billingRepo.CreateCheckout(ctx, input, cashierUserID)
	if err != nil {
		return nil, err
	}

	if s.auditRecorder != nil {
		meta := domain.GetAuditMetadata(ctx)
		var actorID *string
		if cashierUserID != "" {
			actorID = &cashierUserID
		}
		details, _ := json.Marshal(map[string]any{
			"appointment_id":  input.AppointmentID,
			"amount_in_cents": billing.AmountInCents,
			"payment_method":  input.PaymentMethod,
		})
		_ = s.auditRecorder.Record(ctx, domain.AuditLog{
			UnitID:       &billing.UnitID,
			ActorID:      actorID,
			Action:       "CHECKOUT_PROCESSED",
			TargetEntity: "billings",
			TargetID:     billing.ID,
			Details:      details,
			IPAddress:    meta.IPAddress,
			UserAgent:    meta.UserAgent,
		})
	}

	return billing, nil
}
