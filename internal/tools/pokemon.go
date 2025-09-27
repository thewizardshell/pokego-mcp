// Package tools implements MCP tools for Pokemon data retrieval.
// This package provides functionality to interact with the PokeAPI
// and expose Pokemon-related capabilities through the MCP protocol.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"mcp-pokemon/internal/server"
)

// SetupTools configures and registers all Pokemon-related MCP tools.
// It adds three main tools: get_pokemon, list_pokemon, and get_pokemon_details.
func SetupTools(mcp *server.MCPServer) error {
	if mcp == nil || mcp.Server == nil {
		return fmt.Errorf("MCP server is not initialized")
	}

	mcp.Server.AddTool(
		mcpgo.NewTool("get_pokemon",
			mcpgo.WithDescription("Get basic Pokemon information by name or ID"),
			mcpgo.WithString("pokemon", mcpgo.Required()),
		),
		handleGetPokemon,
	)

	mcp.Server.AddTool(
		mcpgo.NewTool("list_pokemon",
			mcpgo.WithDescription("List Pokemon with pagination support"),
			mcpgo.WithNumber("limit", mcpgo.DefaultNumber(20)),
			mcpgo.WithNumber("offset", mcpgo.DefaultNumber(0)),
		),
		handleListPokemon,
	)

	mcp.Server.AddTool(
		mcpgo.NewTool("get_pokemon_details",
			mcpgo.WithDescription("Get detailed Pokemon information including types, abilities, stats, and sprites"),
			mcpgo.WithString("pokemon", mcpgo.Required()),
		),
		handleGetPokemonDetails,
	)

	return nil
}

// handleGetPokemon retrieves basic Pokemon information from PokeAPI.
// It accepts a Pokemon name or ID and returns basic info like name and ID.
func handleGetPokemon(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	pokemon, err := req.RequireString("pokemon")
	if err != nil {
		return mcpgo.NewToolResultError("Missing required parameter 'pokemon'"), nil
	}

	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemon)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return mcpgo.NewToolResultError(fmt.Sprintf("Error creating request: %v", err)), nil
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return mcpgo.NewToolResultError(fmt.Sprintf("Error making request: %v", err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return mcpgo.NewToolResultError(fmt.Sprintf("Pokemon '%s' not found", pokemon)), nil
	}

	body, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	json.Unmarshal(body, &data)

	result := fmt.Sprintf(`{"name": "%s", "id": %.0f}`, data["name"], data["id"])

	return mcpgo.NewToolResultText(result), nil
}

// handleListPokemon retrieves a paginated list of Pokemon from PokeAPI.
// It supports limit and offset parameters for pagination.
func handleListPokemon(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	limit := int(req.GetInt("limit", 20))
	offset := int(req.GetInt("offset", 0))

	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon?limit=%d&offset=%d", limit, offset)
	resp, err := http.Get(url)
	if err != nil {
		return mcpgo.NewToolResultError(fmt.Sprintf("Error making request: %v", err)), nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	return mcpgo.NewToolResultText(string(body)), nil
}

// handleGetPokemonDetails retrieves comprehensive Pokemon information from PokeAPI.
// It returns detailed data including types, abilities, base stats, sprites, and more.
func handleGetPokemonDetails(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	pokemon, err := req.RequireString("pokemon")
	if err != nil {
		return mcpgo.NewToolResultError("Missing required parameter 'pokemon'"), nil
	}

	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemon)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return mcpgo.NewToolResultError(fmt.Sprintf("Error creating request: %v", err)), nil
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return mcpgo.NewToolResultError(fmt.Sprintf("Error making request: %v", err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return mcpgo.NewToolResultError(fmt.Sprintf("Pokemon '%s' not found", pokemon)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return mcpgo.NewToolResultError(fmt.Sprintf("Error reading response: %v", err)), nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return mcpgo.NewToolResultError(fmt.Sprintf("Error parsing JSON: %v", err)), nil
	}

	result := map[string]interface{}{
		"name":   data["name"],
		"id":     data["id"],
		"height": data["height"],
		"weight": data["weight"],
	}

	// Extract types
	if types, ok := data["types"].([]interface{}); ok {
		var typesList []string
		for _, t := range types {
			if typeMap, ok := t.(map[string]interface{}); ok {
				if typeInfo, ok := typeMap["type"].(map[string]interface{}); ok {
					if typeName, ok := typeInfo["name"].(string); ok {
						typesList = append(typesList, typeName)
					}
				}
			}
		}
		result["types"] = typesList
	}

	// Extract abilities
	if abilities, ok := data["abilities"].([]interface{}); ok {
		var abilitiesList []map[string]interface{}
		for _, a := range abilities {
			if abilityMap, ok := a.(map[string]interface{}); ok {
				ability := map[string]interface{}{}
				if abilityInfo, ok := abilityMap["ability"].(map[string]interface{}); ok {
					ability["name"] = abilityInfo["name"]
				}
				if isHidden, ok := abilityMap["is_hidden"].(bool); ok {
					ability["hidden"] = isHidden
				}
				abilitiesList = append(abilitiesList, ability)
			}
		}
		result["abilities"] = abilitiesList
	}

	// Extract base stats
	if stats, ok := data["stats"].([]interface{}); ok {
		baseStats := map[string]interface{}{}
		for _, s := range stats {
			if statMap, ok := s.(map[string]interface{}); ok {
				if statInfo, ok := statMap["stat"].(map[string]interface{}); ok {
					if statName, ok := statInfo["name"].(string); ok {
						if baseStat, ok := statMap["base_stat"].(float64); ok {
							baseStats[statName] = baseStat
						}
					}
				}
			}
		}
		result["base_stats"] = baseStats
	}

	// Extract sprites
	if sprites, ok := data["sprites"].(map[string]interface{}); ok {
		spritesMap := map[string]interface{}{}
		if frontDefault, ok := sprites["front_default"]; ok && frontDefault != nil {
			spritesMap["front_default"] = frontDefault
		}
		if backDefault, ok := sprites["back_default"]; ok && backDefault != nil {
			spritesMap["back_default"] = backDefault
		}
		if frontShiny, ok := sprites["front_shiny"]; ok && frontShiny != nil {
			spritesMap["front_shiny"] = frontShiny
		}
		result["sprites"] = spritesMap
	}

	// Extract base experience
	if baseExperience, ok := data["base_experience"]; ok {
		result["base_experience"] = baseExperience
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcpgo.NewToolResultError(fmt.Sprintf("Error generating result: %v", err)), nil
	}

	return mcpgo.NewToolResultText(string(resultJSON)), nil
}
