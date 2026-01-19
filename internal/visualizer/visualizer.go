package visualizer

// Node represents a graph node
type Node struct {
	ID    string                 `json:"id"`
	Label string                 `json:"label"`
	Type  string                 `json:"type"`
	Data  map[string]interface{} `json:"data"`
}

// Edge represents a graph edge
type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
	Type   string `json:"type"`
}

// GraphVisualizer visualizes graphs
type GraphVisualizer struct {
	nodes map[string]Node
	edges []Edge
}

// NewGraphVisualizer creates a new graph visualizer
func NewGraphVisualizer() *GraphVisualizer {
	return &GraphVisualizer{
		nodes: make(map[string]Node),
		edges: make([]Edge, 0),
	}
}

// AddNode adds a node to the graph
func (v *GraphVisualizer) AddNode(node Node) {
	v.nodes[node.ID] = node
}

// AddEdge adds an edge to the graph
func (v *GraphVisualizer) AddEdge(edge Edge) {
	v.edges = append(v.edges, edge)
}

// GetNodes returns all nodes
func (v *GraphVisualizer) GetNodes() map[string]Node {
	return v.nodes
}

// GetEdges returns all edges
func (v *GraphVisualizer) GetEdges() []Edge {
	return v.edges
}
