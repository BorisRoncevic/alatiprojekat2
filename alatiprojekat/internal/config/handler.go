package	config
import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ConfigHandler struct {
	service *ConfigService
}

func NewConfigHandler(service *ConfigService) *ConfigHandler {
	return &ConfigHandler{service: service}
}

//hendleri za konfiguraciju

func (h *ConfigHandler) CreateConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	var config Configuration
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	createdConfig, err := h.service.CreateConfiguration(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdConfig)
}

func (h *ConfigHandler) GetConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	version := vars["version"]

	config, err := h.service.GetConfiguration(id, version)
	if err != nil {
		if strings.Contains(err.Error(), "nije pronađena") || strings.Contains(err.Error(), "ključ") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (h *ConfigHandler) DeleteConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	version := vars["version"]

	if err := h.service.DeleteConfiguration(id, version); err != nil {
		if strings.Contains(err.Error(), "nije pronađena") {
			http.Error(w, err.Error(), http.StatusNotFound) // Vraća 404
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

//hendler za pretragu
func (h *ConfigHandler) SearchConfigurationsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	labelsToSearch := make(map[string]string)
	for key, values := range query {
		if len(values) > 0 {
			labelsToSearch[key] = values[0]
		}
	}

	if len(labelsToSearch) == 0 {
		http.Error(w, "Niste uneli nijednu labelu za pretragu", http.StatusBadRequest)
		return
	}

	configs, err := h.service.SearchConfigurationsByLabels(labelsToSearch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configs)
}


// UpdateConfigurationHandler azuriranje postojece konfiguracije
func (h *ConfigHandler) UpdateConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]
	version := vars["version"]

	var config Configuration
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//provera id
	if config.ID != id || config.Version != version {
		http.Error(w, "ID ili verzija u telu zahteva se ne poklapaju sa onima u URL-u", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateConfiguration(config); err != nil {
		if strings.Contains(err.Error(), "nije pronađena") {
			http.Error(w, err.Error(), http.StatusNotFound) // Vraća 404
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(config)
}


func (h *ConfigHandler) CreateConfigurationGroupHandler(w http.ResponseWriter, r *http.Request) {
	var group ConfigurationGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	group.ID = uuid.New().String()
	createdGroup, err := h.service.CreateConfigurationGroup(group)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdGroup)
}

func (h *ConfigHandler) GetConfigurationGroupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	version := vars["version"]
	group, err := h.service.GetConfigurationGroup(id, version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

func (h *ConfigHandler) DeleteConfigurationGroupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	version := vars["version"]
	if err := h.service.DeleteConfigurationGroup(id, version); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}


// SearchConfigurationGroupsHandler GET pretraga
func (h *ConfigHandler) SearchConfigurationGroupsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	labelsToSearch := make(map[string]string)
	for key, values := range query {
		if len(values) > 0 {
			labelsToSearch[key] = values[0]
		}
	}

	if len(labelsToSearch) == 0 {
		http.Error(w, "Niste uneli nijednu labelu za pretragu grupe", http.StatusBadRequest)
		return
	}
}

func (h *ConfigHandler) UpdateConfigurationGroupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]
	version := vars["version"]

	var group ConfigurationGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Provera konzistentnosti ID-jeva
	if group.ID != id || group.Version != version {
		http.Error(w, "ID ili verzija grupe u telu zahteva se ne poklapaju sa onima u URL-u", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateConfigurationGroup(group); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(group)
}
