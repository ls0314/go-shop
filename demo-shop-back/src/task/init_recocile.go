package task

import "time"

func Init() {
	reconcile := NewReconcileService()
	go reconcile.Start(5*time.Minute, 12*time.Hour)
	stockReconcile := NewStockReconcileService()
	go stockReconcile.Start(5 * time.Minute)
}
