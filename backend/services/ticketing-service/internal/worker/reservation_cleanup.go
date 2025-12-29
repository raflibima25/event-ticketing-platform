package worker

import (
	"context"
	"log"
	"time"

	"github.com/raflibima25/event-ticketing-platform/backend/services/ticketing-service/internal/service"
)

// ReservationCleanupWorker handles periodic cleanup of expired reservations
type ReservationCleanupWorker struct {
	reservationService service.ReservationService
	interval           time.Duration
	stopChan           chan struct{}
}

// NewReservationCleanupWorker creates new cleanup worker instance
func NewReservationCleanupWorker(
	reservationService service.ReservationService,
	interval time.Duration,
) *ReservationCleanupWorker {
	return &ReservationCleanupWorker{
		reservationService: reservationService,
		interval:           interval,
		stopChan:           make(chan struct{}),
	}
}

// Start begins the cleanup worker
func (w *ReservationCleanupWorker) Start(ctx context.Context) {
	log.Printf("[Worker] Reservation cleanup worker started (interval: %v)", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run cleanup immediately on start
	w.runCleanup(ctx)

	for {
		select {
		case <-ticker.C:
			w.runCleanup(ctx)
		case <-w.stopChan:
			log.Println("[Worker] Reservation cleanup worker stopped")
			return
		case <-ctx.Done():
			log.Println("[Worker] Reservation cleanup worker stopped due to context cancellation")
			return
		}
	}
}

// Stop gracefully stops the cleanup worker
func (w *ReservationCleanupWorker) Stop() {
	close(w.stopChan)
}

// runCleanup executes the cleanup operation
func (w *ReservationCleanupWorker) runCleanup(ctx context.Context) {
	log.Println("[Worker] Running reservation cleanup...")

	startTime := time.Now()
	count, err := w.reservationService.CleanupExpiredReservations(ctx)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("[Worker] Cleanup failed: %v (duration: %v)", err, duration)
		return
	}

	if count > 0 {
		log.Printf("[Worker] Cleanup completed: %d expired reservations released (duration: %v)", count, duration)
	} else {
		log.Printf("[Worker] Cleanup completed: no expired reservations found (duration: %v)", duration)
	}
}
