package service

import (
	"context"
	"testing"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

type mockBillingRepo struct {
	checkoutFn func(ctx context.Context, input domain.CheckoutInput, cashierUserID string) (*domain.Billing, error)
}

func (m *mockBillingRepo) CreateCheckout(ctx context.Context, input domain.CheckoutInput, cashierUserID string) (*domain.Billing, error) {
	if m.checkoutFn != nil {
		return m.checkoutFn(ctx, input, cashierUserID)
	}
	return &domain.Billing{
		ID:            "bill-1",
		AppointmentID: &input.AppointmentID,
		PaymentMethod: input.PaymentMethod,
		Status:        domain.BillingStatusPaid,
		CreatedAt:     time.Now(),
	}, nil
}

func TestBillingService_Checkout(t *testing.T) {
	repo := &mockBillingRepo{}
	svc := NewBillingService(repo)
	ctx := context.Background()

	t.Run("successfully processes checkout with PIX", func(t *testing.T) {
		bill, err := svc.Checkout(ctx, domain.CheckoutInput{
			AppointmentID: "apt-1",
			PaymentMethod: domain.PaymentMethodPix,
		}, "cashier-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bill.PaymentMethod != domain.PaymentMethodPix {
			t.Errorf("expected PIX, got %s", bill.PaymentMethod)
		}
	})

	t.Run("rejects invalid payment method", func(t *testing.T) {
		_, err := svc.Checkout(ctx, domain.CheckoutInput{
			AppointmentID: "apt-1",
			PaymentMethod: "BITCOIN",
		}, "cashier-1")
		if err != domain.ErrInvalidPaymentMethod {
			t.Fatalf("expected ErrInvalidPaymentMethod, got %v", err)
		}
	})

	t.Run("rejects empty appointment id", func(t *testing.T) {
		_, err := svc.Checkout(ctx, domain.CheckoutInput{
			AppointmentID: "",
			PaymentMethod: domain.PaymentMethodCash,
		}, "cashier-1")
		if err != domain.ErrAppointmentNotFound {
			t.Fatalf("expected ErrAppointmentNotFound, got %v", err)
		}
	})
}
