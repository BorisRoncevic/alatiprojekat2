package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/BorisRoncevic/alatiprojekat2/internal/config"
	"github.com/BorisRoncevic/alatiprojekat2/internal/middleware"
	"golang.org/x/time/rate"
)

func main() {
	router := mux.NewRouter()
	configService := config.NewConfigService()
	configHandler := config.NewConfigHandler(configService)
	limiter := rate.NewLimiter(1, 3)

	// --- KONFIGURACIJE ---
	router.HandleFunc("/configs", configHandler.CreateConfigurationHandler).Methods("POST")

	// ISPRAVLJENE RUTE: Koriste /configs/{id}/{version}
	router.HandleFunc("/configs/{id}/{version}", configHandler.GetConfigurationHandler).Methods("GET")
	router.HandleFunc("/configs/{id}/{version}", configHandler.UpdateConfigurationHandler).Methods("PUT")
	router.HandleFunc("/configs/{id}/{version}", configHandler.DeleteConfigurationHandler).Methods("DELETE")

	// RUTA ZA PRETRAGU KONFIGURACIJA
	router.HandleFunc("/configs/search", configHandler.SearchConfigurationsHandler).Methods("GET")

	// --- GRUPE KONFIGURACIJA ---
	router.HandleFunc("/groups", configHandler.CreateConfigurationGroupHandler).Methods("POST")

	// ISPRAVLJENE RUTE: Koriste /groups/{id}/{version}
	router.HandleFunc("/groups/{id}/{version}", configHandler.GetConfigurationGroupHandler).Methods("GET")
	router.HandleFunc("/groups/{id}/{version}", configHandler.UpdateConfigurationGroupHandler).Methods("PUT")
	router.HandleFunc("/groups/{id}/{version}", configHandler.DeleteConfigurationGroupHandler).Methods("DELETE")

	// RUTA ZA PRETRAGU GRUPA
	router.HandleFunc("/groups/search", configHandler.SearchConfigurationGroupsHandler).Methods("GET")

	srv := &http.Server{
		Addr:    ":8080",
		Handler: middleware.RateLimitMiddleware(limiter)(router),
	}

	go func() {
		log.Println("Server startovan na portu 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Greška pri pokretanju servera: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Server se gasi...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server je nasilno ugašen:", err)
	}

	log.Println("Server je uspešno ugašen")
}
