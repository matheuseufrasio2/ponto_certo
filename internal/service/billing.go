package service

import (
	"context"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/repository"
)

// BillingService manages billing and financial transactions.
type BillingService struct {
	billingRepo repository.BillingRepository
}

// NewBillingService instantiates BillingService.
func NewBillingService(billingRepo repository.BillingRepository) *BillingService {
	return &BillingService{billingRepo: billingRepo}
}

// Checkout executes final payment processing for an appointment.
func (s *BillingService) Checkout(ctx context.Context, input domain.CheckoutInput, cashierUserID string) (*domain.Billing, error) {
	if input.AppointmentID == "" {
		return nil, domain.ErrAppointmentNotFound
	}
	if !input.PaymentMethod.IsValid() {
		return nil, domain.ErrInvalidPaymentMethod
	}

	return s.billingRepo.CreateCheckout(ctx, input, cashierUserID)
}
