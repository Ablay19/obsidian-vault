package visualizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGraphVisualizer(t *testing.T) {
	v := NewGraphVisualizer()

	assert.NotNil(t, v)
	assert.NotNil(t, v.nodes)
	assert.NotNil(t, v.edges)
	assert.Equal(t, 0, len(v.nodes))
	assert.Equal(t, 0, len(v.edges))
}

func TestGraphVisualizer_AddNode(t *testing.T) {
	v := NewGraphVisualizer()

	node := Node{
		ID:    "test_node",
		Label: "Test Node",
		Type:  "test",
		Data: map[string]interface{}{
			"key": "value",
		},
	}

	v.AddNode(node)

	assert.Equal(t, 1, len(v.nodes))
	assert.Equal(t, node, v.nodes["test_node"])
}

func TestGraphVisualizer_AddEdge(t *testing.T) {
	v := NewGraphVisualizer()

	edge := Edge{
		Source: "node1",
		Target: "node2",
		Label:  "connects",
		Type:   "relationship",
	}

	v.AddEdge(edge)

	assert.Equal(t, 1, len(v.edges))
	assert.Contains(t, v.edges, edge)
}

func TestGraphVisualizer_GetNodes(t *testing.T) {
	v := NewGraphVisualizer()

	node1 := Node{ID: "node1", Label: "Node 1"}
	node2 := Node{ID: "node2", Label: "Node 2"}

	v.AddNode(node1)
	v.AddNode(node2)

	nodes := v.GetNodes()

	assert.Equal(t, 2, len(nodes))
	assert.Contains(t, nodes, node1)
	assert.Contains(t, nodes, node2)
}

func TestGraphVisualizer_GetEdges(t *testing.T) {
	v := NewGraphVisualizer()

	edge1 := Edge{Source: "node1", Target: "node2", Label: "edge1"}
	edge2 := Edge{Source: "node2", Target: "node3", Label: "edge2"}

	v.AddEdge(edge1)
	v.AddEdge(edge2)

	edges := v.GetEdges()

	assert.Equal(t, 2, len(edges))
	assert.Contains(t, edges, edge1)
	assert.Contains(t, edges, edge2)
}

func TestGraphVisualizer_GetNodeByID(t *testing.T) {
	v := NewGraphVisualizer()

	node := Node{ID: "test_node", Label: "Test Node"}
	v.AddNode(node)

	retrieved, exists := v.GetNodeByID("test_node")

	assert.True(t, exists)
	assert.Equal(t, node, retrieved)
}

func TestGraphVisualizer_GetNodeByID_NotFound(t *testing.T) {
	v := NewGraphVisualizer()

	retrieved, exists := v.GetNodeByID("nonexistent")

	assert.False(t, exists)
	assert.Equal(t, Node{}, retrieved)
}

func TestGraphVisualizer_GetEdgesForNode(t *testing.T) {
	v := NewGraphVisualizer()

	edge1 := Edge{Source: "node1", Target: "node2", Label: "edge1"}
	edge2 := Edge{Source: "node1", Target: "node3", Label: "edge2"}
	edge3 := Edge{Source: "node2", Target: "node1", Label: "edge3"}

	v.AddEdge(edge1)
	v.AddEdge(edge2)
	v.AddEdge(edge3)

	edges := v.GetEdgesForNode("node1")

	assert.Equal(t, 3, len(edges))
	assert.Contains(t, edges, edge1)
	assert.Contains(t, edges, edge2)
	assert.Contains(t, edges, edge3)
}

func TestGraphVisualizer_GetNeighbors(t *testing.T) {
	v := NewGraphVisualizer()

	// Add nodes
	node1 := Node{ID: "node1", Label: "Node 1"}
	node2 := Node{ID: "node2", Label: "Node 2"}
	node3 := Node{ID: "node3", Label: "Node 3"}

	v.AddNode(node1)
	v.AddNode(node2)
	v.AddNode(node3)

	// Add edges
	v.AddEdge(Edge{Source: "node1", Target: "node2"})
	v.AddEdge(Edge{Source: "node1", Target: "node3"})
	v.AddEdge(Edge{Source: "node2", Target: "node3"})

	neighbors := v.GetNeighbors("node1")

	assert.Equal(t, 2, len(neighbors))
	assert.Contains(t, neighbors, node2)
	assert.Contains(t, neighbors, node3)
}

func TestGraphVisualizer_RemoveNode(t *testing.T) {
	v := NewGraphVisualizer()

	node := Node{ID: "test_node", Label: "Test Node"}
	v.AddNode(node)

	assert.Equal(t, 1, len(v.nodes))

	v.RemoveNode("test_node")

	assert.Equal(t, 0, len(v.nodes))
}

func TestGraphVisualizer_RemoveEdge(t *testing.T) {
	v := NewGraphVisualizer()

	edge := Edge{Source: "node1", Target: "node2", Label: "test_edge"}
	v.AddEdge(edge)

	assert.Equal(t, 1, len(v.edges))

	v.RemoveEdge(edge)

	assert.Equal(t, 0, len(v.edges))
}

func TestGraphVisualizer_Clear(t *testing.T) {
	v := NewGraphVisualizer()

	v.AddNode(Node{ID: "node1"})
	v.AddNode(Node{ID: "node2"})
	v.AddEdge(Edge{Source: "node1", Target: "node2"})

	assert.Equal(t, 2, len(v.nodes))
	assert.Equal(t, 1, len(v.edges))

	v.Clear()

	assert.Equal(t, 0, len(v.nodes))
	assert.Equal(t, 0, len(v.edges))
}

func TestGraphVisualizer_GetStats(t *testing.T) {
	v := NewGraphVisualizer()

	v.AddNode(Node{ID: "node1", Type: "type1"})
	v.AddNode(Node{ID: "node2", Type: "type1"})
	v.AddNode(Node{ID: "node3", Type: "type2"})
	v.AddEdge(Edge{Source: "node1", Target: "node2", Type: "relationship1"})
	v.AddEdge(Edge{Source: "node2", Target: "node3", Type: "relationship2"})

	stats := v.GetStats()

	assert.Equal(t, 3, stats["total_nodes"])
	assert.Equal(t, 2, stats["total_edges"])
	assert.Equal(t, 2, stats["node_types"])
	assert.Equal(t, 2, stats["edge_types"])
}

func TestGraphVisualizer_ToJSON(t *testing.T) {
	v := NewGraphVisualizer()

	node := Node{ID: "test_node", Label: "Test Node", Type: "test"}
	v.AddNode(node)

	json := v.ToJSON()

	assert.NotEmpty(t, json)
	assert.Contains(t, json, "test_node")
	assert.Contains(t, json, "Test Node")
}

func TestGraphVisualizer_FromJSON(t *testing.T) {
	jsonData := `{
		"nodes": [
			{"id": "node1", "label": "Node 1", "type": "test"}
		],
		"edges": [
			{"source": "node1", "target": "node2", "label": "connects"}
		]
	}`

	v := NewGraphVisualizer()
	err := v.FromJSON(jsonData)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(v.nodes))
	assert.Equal(t, 1, len(v.edges))
}

func TestGraphVisualizer_FindPath(t *testing.T) {
	v := NewGraphVisualizer()

	// Add nodes
	v.AddNode(Node{ID: "A"})
	v.AddNode(Node{ID: "B"})
	v.AddNode(Node{ID: "C"})
	v.AddNode(Node{ID: "D"})

	// Add edges: A -> B -> C -> D
	v.AddEdge(Edge{Source: "A", Target: "B"})
	v.AddEdge(Edge{Source: "B", Target: "C"})
	v.AddEdge(Edge{Source: "C", Target: "D"})

	path := v.FindPath("A", "D")

	assert.NotNil(t, path)
	assert.Equal(t, 4, len(path))
	assert.Equal(t, "A", path[0])
	assert.Equal(t, "B", path[1])
	assert.Equal(t, "C", path[2])
	assert.Equal(t, "D", path[3])
}

func TestGraphVisualizer_FindPath_NoPath(t *testing.T) {
	v := NewGraphVisualizer()

	v.AddNode(Node{ID: "A"})
	v.AddNode(Node{ID: "B"})
	v.AddNode(Node{ID: "C"})

	// No connection between A and C
	v.AddEdge(Edge{Source: "A", Target: "B"})

	path := v.FindPath("A", "C")

	assert.Nil(t, path)
}

func TestGraphVisualizer_GetConnectedComponents(t *testing.T) {
	v := NewGraphVisualizer()

	// Component 1: A -> B -> C
	v.AddNode(Node{ID: "A"})
	v.AddNode(Node{ID: "B"})
	v.AddNode(Node{ID: "C"})
	v.AddEdge(Edge{Source: "A", Target: "B"})
	v.AddEdge(Edge{Source: "B", Target: "C"})

	// Component 2: D -> E
	v.AddNode(Node{ID: "D"})
	v.AddNode(Node{ID: "E"})
	v.AddEdge(Edge{Source: "D", Target: "E"})

	// Isolated node
	v.AddNode(Node{ID: "F"})

	components := v.GetConnectedComponents()

	assert.Equal(t, 3, len(components))

	// Find components by size
	var componentSizes []int
	for _, component := range components {
		componentSizes = append(componentSizes, len(component))
	}

	assert.Contains(t, componentSizes, 3) // A, B, C
	assert.Contains(t, componentSizes, 2) // D, E
	assert.Contains(t, componentSizes, 1) // F
}
