package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/5amCurfew/orca/lib"
	"github.com/gin-gonic/gin"
)

type GraphJSON struct {
	Tasks    map[string]*lib.Task `json:"tasks"`
	Nodes    map[string]struct{}  `json:"nodes"`
	Parents  map[string][]string  `json:"parents"`
	Children map[string][]string  `json:"children"`
}

func convertToGraphJSON(input lib.Graph) GraphJSON {
	out := GraphJSON{}
	out.Tasks = input.Tasks
	out.Nodes = input.Nodes
	out.Parents = make(map[string][]string)
	out.Children = make(map[string][]string)

	for key := range input.Parents {
		out.Parents[key] = mapToJSONArray(input.Parents[key])
	}

	for key := range input.Children {
		out.Children[key] = mapToJSONArray(input.Children[key])
	}

	return out
}

func mapToJSONArray(input map[string]struct{}) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	return keys
}

func Graph(c *gin.Context) {
	var requestData map[string]interface{}

	// Parse JSON request body
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract orca file path from the request
	filePath, ok := requestData["file_path"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_path required"})
		return
	}

	g, err := lib.NewGraph(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to parse DAG: %s", err)})
		return
	}

	gJSON := convertToGraphJSON(*g)
	jsonRepresentation, _ := json.MarshalIndent(gJSON, "", "  ")
	fmt.Println(string(jsonRepresentation))

	c.JSON(http.StatusOK, gin.H{
		"graph":   json.RawMessage(jsonRepresentation),
		"message": fmt.Sprintf("DAG %s graph created", filePath),
	})
}
