package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const agentsURL = "https://valorant-api.com/v1/agents?isPlayableCharacter=true"

// DefaultAgentNames provides built-in fallback mapping for Valorant agents.
var DefaultAgentNames = map[string]string{
	"dade69b4-4f5a-8528-247b-219e5a1facd6": "Fade",
	"5f8d3d7f-467b-97f3-062c-13acf203c006": "Breach",
	"f94c3b30-42be-e959-889c-5aa313dba261": "Raze",
	"22697a3d-45bf-8dd7-4fec-84a9e28c69d7": "Chamber",
	"601dbbe7-43ce-be57-2a40-4abd24953621": "KAY/O",
	"6f2a04ca-43e0-be17-7f36-b39086d3548b": "Skye",
	"117ed9e3-49f3-6512-3ccf-0cada7e3823b": "Cypher",
	"ded3520f-4264-bfed-162d-b080e2abccf9": "Sova",
	"320799f3-4048-5770-8757-7cf7b092acb6": "Viper",
	"707eab51-4727-4cb4-3235-e693e4d60849": "Phoenix",
	"eb9333ab-4034-4c81-1a60-d581175929f8": "Yoru",
	"41fb69c1-411a-7b37-be4a-da8671699d30": "Killjoy",
	"9f0d8ba9-42c6-9247-3c93-2abba5d77ccc": "Omen",
	"7f94d92c-4234-0a36-9646-3a87eb8b5c89": "Yoru",
	"569fdd95-4d10-43ab-ca70-79becc718b46": "Sage",
	"a3bfb80f-4041-f0a0-a540-49b4e40b00a5": "Reyna",
	"8e252d0a-4ee5-b42d-7e20-77166af5f4ff": "Yoru",
	"add6443a-4814-a636-2241-60a3a2777160": "Jett",
	"b0627010-4071-be69-ed15-00ac0a772b10": "Harbor",
	"e370fa57-4757-3604-3648-499e1f642d3f": "Gekko",
	"cc8e01d3-47f9-6378-2abf-4b7f0a7e739e": "Deadlock",
	"0e38b806-40fc-8844-aca7-20e64405d6b3": "Iso",
	"1e58de9c-4950-523e-f32c-70a380e804ba": "Clove",
	"95b78ed7-4637-86a9-7a47-f2a2b5402499": "Vyse",
	"efba5359-4016-a1e5-7626-b1ae76895940": "Tejo",
}

type agentsCache struct {
	Version  string            `json:"version"`
	Agents   map[string]string `json:"agents"`
	CachedAt time.Time         `json:"cachedAt"`
}

type agentsAPIResponse struct {
	Status int `json:"status"`
	Data   []struct {
		UUID        string `json:"uuid"`
		DisplayName string `json:"displayName"`
	} `json:"data"`
}

// LoadOrFetchAgents returns map from agent UUID -> DisplayName.
func LoadOrFetchAgents(currentVersion string) (map[string]string, error) {
	dir, err := cacheDir()
	if err != nil {
		return DefaultAgentNames, nil
	}
	path := filepath.Join(dir, "agents.json")

	if data, err := os.ReadFile(path); err == nil {
		var ac agentsCache
		if json.Unmarshal(data, &ac) == nil && ac.Agents != nil && len(ac.Agents) > 0 {
			if currentVersion == "" || ac.Version == currentVersion {
				return ac.Agents, nil
			}
		}
	}

	agents, err := fetchAgents()
	if err != nil {
		return DefaultAgentNames, nil
	}

	ac := agentsCache{Version: currentVersion, Agents: agents, CachedAt: time.Now()}
	if data, err := json.MarshalIndent(ac, "", "  "); err == nil {
		_ = os.WriteFile(path, data, 0o644)
	}

	return agents, nil
}

func fetchAgents() (map[string]string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(agentsURL)
	if err != nil {
		return nil, fmt.Errorf("fetch agents: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ar agentsAPIResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return nil, err
	}

	agents := make(map[string]string, len(ar.Data))
	for _, a := range ar.Data {
		agents[strings.ToLower(a.UUID)] = a.DisplayName
	}

	if len(agents) == 0 {
		return DefaultAgentNames, nil
	}
	return agents, nil
}

// GetAgentName returns display name for an agent UUID.
func GetAgentName(uuid string, agentsMap map[string]string) string {
	uuidLower := strings.ToLower(uuid)
	if agentsMap != nil {
		if name, ok := agentsMap[uuidLower]; ok && name != "" {
			return name
		}
	}
	if name, ok := DefaultAgentNames[uuidLower]; ok {
		return name
	}
	return "Agent"
}
