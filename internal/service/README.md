# Service Layer (internal/service)

The **Service** layer is responsible for the **business logic and use cases** of the barbershop.

## Idiomatic Go Principles:
1. **Consumer-Defined Interfaces**: The service defines the repository interface it requires, rather than importing interfaces from the domain or repository packages:
   ```go
   type AppointmentRepository interface {
       Create(ctx context.Context, appt domain.Appointment) error
       HasConflict(ctx context.Context, barberID string, start, end time.Time) (bool, error)
   }
   ```
2. **Constructor Dependency Injection**:
   ```go
   type AppointmentService struct {
       repo AppointmentRepository
   }

   func NewAppointmentService(repo AppointmentRepository) *AppointmentService {
       return &AppointmentService{repo: repo}
   }
   ```
3. **Typical Use Cases**:
   - Check barber availability for requested appointment time slots.
   - Calculate appointment end time (`EndTime = StartTime + Service.DurationMinutes`).
   - Trigger customer confirmation notifications or webhook dispatches.
