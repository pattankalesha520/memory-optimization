package main
import (
	"fmt"
	"math/rand"
	"time"
)
type Node struct {
	id         int
	totalMemMB int
	usedMemMB  int
}
func createNodes(count int) []*Node {
	nodes := make([]*Node, count)
	for i := range nodes {
		nodes[i] = &Node{
			id:         i + 1,
			totalMemMB: 1024 + rand.Intn(1024),
			usedMemMB:  rand.Intn(512),
		}
	}
	return nodes
}
func simulateMonitoringAgent(node *Node) int {
	time.Sleep(time.Millisecond * time.Duration(10+rand.Intn(10)))
	node.usedMemMB += rand.Intn(60) - 30
	if node.usedMemMB < 0 {
		node.usedMemMB = 0
	}
	if node.usedMemMB > node.totalMemMB {
		node.usedMemMB = node.totalMemMB
	}
	return node.usedMemMB
}
func ruleBasedController(node *Node) string {
	const (
		scaleUpThreshold   = 80
		scaleDownThreshold = 40
	)
	usagePercent := float64(node.usedMemMB) / float64(node.totalMemMB) * 100
	switch {
	case usagePercent > scaleUpThreshold:
		return "SCALE_UP"
	case usagePercent < scaleDownThreshold:
		return "SCALE_DOWN"
	default:
		return "STABLE"
	}
}
func autoscaler(node *Node, decision string) {
	switch decision {
	case "SCALE_UP":
		node.totalMemMB += 256
	case "SCALE_DOWN":
		if node.totalMemMB > 512 {
			node.totalMemMB -= 128
		}
	}
	time.Sleep(time.Millisecond * 5)
}
func runCycle(nodes []*Node) {
	for _, n := range nodes {
		memUsage := simulateMonitoringAgent(n)
		decision := ruleBasedController(n)
		autoscaler(n, decision)
		fmt.Printf("Node:%02d | Total:%4dMB | Used:%4dMB | Action:%-10s\n",
			n.id, n.totalMemMB, memUsage, decision)
	}
}
func main() {
	rand.Seed(time.Now().UnixNano())
	nodes := createNodes(5)
	fmt.Println("=== Legacy Memory Optimization Simulation ===")
	for cycle := 1; cycle <= 5; cycle++ {
		fmt.Printf("\nCycle %d:\n", cycle)
		runCycle(nodes)
		time.Sleep(time.Second)
	}
}
