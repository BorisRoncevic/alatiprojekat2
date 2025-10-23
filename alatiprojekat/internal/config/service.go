package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/consul/api"
)

// ConfigService koristi Consul klijenta za rad sa konfiguracijama
type ConfigService struct {
	consulClient *api.Client
}

// NewConfigService inicijalizuje i vraća novi ConfigService sa povezanim Consul klijentom
func NewConfigService() *ConfigService {
	config := api.DefaultConfig()
	config.Address = os.Getenv("CONSUL_ADDRESS")
	if config.Address == "" {
		config.Address = "127.0.0.1:8500" //adresa Consula
	}

	client, err := api.NewClient(config)
	if err != nil {
		panic(fmt.Sprintf("Greška pri inicijalizaciji Consul klijenta: %v", err))
	}

	return &ConfigService{consulClient: client}
}

func (s *ConfigService) putToConsul(key string, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("greška pri serijalizaciji podataka: %w", err)
	}

	p := &api.KVPair{Key: key, Value: bytes}
	_, err = s.consulClient.KV().Put(p, nil)
	return err
}

// citamo kljuc-vrednost iz Consula.
func (s *ConfigService) getFromConsul(key string) (*api.KVPair, error) {
	pair, _, err := s.consulClient.KV().Get(key, nil)
	if err != nil {
		return nil, fmt.Errorf("greška pri čitanju iz Consula: %w", err)
	}
	if pair == nil {
		return nil, fmt.Errorf("ključ %s nije pronađen", key)
	}
	return pair, nil
}
func (s *ConfigService) CreateConfiguration(config Configuration) (Configuration, error) {
	key := fmt.Sprintf("configs/%s/%s", config.ID, config.Version)

	_, err := s.getFromConsul(key)
	if err == nil {
		return Configuration{}, fmt.Errorf("konfiguracija sa ID-jem %s i verzijom %s već postoji", config.ID, config.Version)
	}

	// cuuvanje u Consulu
	if err := s.putToConsul(key, config); err != nil {
		return Configuration{}, fmt.Errorf("greška pri čuvanju konfiguracije u Consulu: %w", err)
	}
	return config, nil
}

func (s *ConfigService) GetConfiguration(id, version string) (Configuration, error) {
	key := fmt.Sprintf("configs/%s/%s", id, version)
	pair, err := s.getFromConsul(key)
	if err != nil {
		return Configuration{}, fmt.Errorf("konfiguracija sa ID-jem %s i verzijom %s nije pronađena", id, version)
	}

	var config Configuration
	if err := json.Unmarshal(pair.Value, &config); err != nil {
		return Configuration{}, fmt.Errorf("greška pri deserijalizaciji konfiguracije: %w", err)
	}
	return config, nil
}

func (s *ConfigService) DeleteConfiguration(id, version string) error {
	key := fmt.Sprintf("configs/%s/%s", id, version)

	if _, err := s.getFromConsul(key); err != nil {
		return fmt.Errorf("konfiguracija sa ID-jem %s i verzijom %s nije pronađena: %w", id, version, err)
	}

	_, err := s.consulClient.KV().Delete(key, nil)
	if err != nil {
		return fmt.Errorf("greška pri brisanju konfiguracije: %w", err)
	}
	return nil
}

func (s *ConfigService) UpdateConfiguration(config Configuration) error {
	key := fmt.Sprintf("configs/%s/%s", config.ID, config.Version)

	if _, err := s.getFromConsul(key); err != nil {
		return fmt.Errorf("konfiguracija sa ID-jem %s i verzijom %s nije pronađena za ažuriranje", config.ID, config.Version)
	}

	if err := s.putToConsul(key, config); err != nil {
		return fmt.Errorf("greška pri ažuriranju konfiguracije u Consulu: %w", err)
	}
	return nil
}

//pretrazivanje konfiguracije na osnovu labele u Consulu
func (s *ConfigService) SearchConfigurationsByLabels(labelsToSearch map[string]string) ([]Configuration, error) {
	pairs, _, err := s.consulClient.KV().List("configs/", nil)
	if err != nil {
		return nil, fmt.Errorf("greška pri preuzimanju lista konfiguracija iz Consula: %w", err)
	}
	if pairs == nil {
		return []Configuration{}, nil
	}

	var result []Configuration
	for _, pair := range pairs {
		var config Configuration
		if err := json.Unmarshal(pair.Value, &config); err != nil {
			fmt.Printf("Greška pri deserijalizaciji: %v. Nastavljam dalje.\n", err)
			continue
		}

		matches := true
		for searchKey, searchValue := range labelsToSearch {
			value, ok := config.Labels[searchKey]
			if !ok || value != searchValue {
				matches = false
				break
			}
		}

		if matches {
			result = append(result, config)
		}
	}
	return result, nil
}

func (s *ConfigService) CreateConfigurationGroup(group ConfigurationGroup) (ConfigurationGroup, error) {
	key := fmt.Sprintf("groups/%s/%s", group.ID, group.Version)

	_, err := s.getFromConsul(key)
	if err == nil {
		return ConfigurationGroup{}, fmt.Errorf("grupa sa ID-jem %s i verzijom %s već postoji", group.ID, group.Version)
	}

	if err := s.putToConsul(key, group); err != nil {
		return ConfigurationGroup{}, fmt.Errorf("greška pri čuvanju grupe u Consulu: %w", err)
	}
	return group, nil
}

//preuzima grupu konfiguracija
func (s *ConfigService) GetConfigurationGroup(id, version string) (ConfigurationGroup, error) {
	key := fmt.Sprintf("groups/%s/%s", id, version)
	pair, err := s.getFromConsul(key)
	if err != nil {
		return ConfigurationGroup{}, fmt.Errorf("grupa sa ID-jem %s i verzijom %s nije pronađena", id, version)
	}

	var group ConfigurationGroup
	if err := json.Unmarshal(pair.Value, &group); err != nil {
		return ConfigurationGroup{}, fmt.Errorf("greška pri deserijalizaciji grupe: %w", err)
	}
	return group, nil
}

func (s *ConfigService) DeleteConfigurationGroup(id, version string) error {
	key := fmt.Sprintf("groups/%s/%s", id, version)

	if _, err := s.getFromConsul(key); err != nil {
		return fmt.Errorf("grupa sa ID-jem %s i verzijom %s nije pronađena: %w", id, version, err)
	}

	_, err := s.consulClient.KV().Delete(key, nil)
	if err != nil {
		return fmt.Errorf("greška pri brisanju grupe: %w", err)
	}
	return nil
}

func (s *ConfigService) UpdateConfigurationGroup(group ConfigurationGroup) error {
	key := fmt.Sprintf("groups/%s/%s", group.ID, group.Version)

	existingPair, _, err := s.consulClient.KV().Get(key, nil)
	if err != nil {
		return fmt.Errorf("greška pri proveri grupe %s:%s u Consulu: %w", group.ID, group.Version, err)
	}

	if existingPair == nil {
		return fmt.Errorf("grupa konfiguracija sa ID-jem %s i verzijom %s nije pronađena za ažuriranje", group.ID, group.Version)
	}

	data, err := json.Marshal(group)
	if err != nil {
		return fmt.Errorf("greška pri serijalizaciji ažurirane grupe: %w", err)
	}

	p := &api.KVPair{Key: key, Value: data}
	if _, err := s.consulClient.KV().Put(p, nil); err != nil {
		return fmt.Errorf("greška pri ažuriranju grupe %s:%s u Consulu: %w", group.ID, group.Version, err)
	}

	return nil
}

func (s *ConfigService) SearchConfigurationGroupsByLabels(labelsToSearch map[string]string) ([]ConfigurationGroup, error) {
	pairs, _, err := s.consulClient.KV().List("groups/", nil)
	if err != nil {
		return nil, fmt.Errorf("greška pri preuzimanju lista grupa konfiguracija iz Consula: %w", err)
	}
	if pairs == nil {
		return []ConfigurationGroup{}, nil
	}

	var result []ConfigurationGroup
	for _, pair := range pairs {
		var group ConfigurationGroup
		if err := json.Unmarshal(pair.Value, &group); err != nil {
			fmt.Printf("Greška pri deserijalizaciji grupe: %v. Nastavljam dalje.\n", err)
			continue
		}

		// pretraga po labelama
		matches := true
		for searchKey, searchValue := range labelsToSearch {
			value, ok := group.Labels[searchKey]
			if !ok || value != searchValue {
				matches = false
				break
			}
		}

		if matches {
			result = append(result, group)
		}
	}
	return result, nil
}
