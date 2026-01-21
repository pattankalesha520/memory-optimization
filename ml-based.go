import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

type Node struct {
	ID, TotalMB, UsedMB int
	mu                  sync.Mutex
}

type FeatureStore struct {
	mu     sync.Mutex
	window int
	data   map[int][]float64
}

func NewFeatureStore(w int) *FeatureStore {
	return &FeatureStore{window: w, data: make(map[int][]float64)}
}

func (fs *FeatureStore) Add(nodeID, used int) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	s := fs.data[nodeID]
	s = append(s, float64(used))
	if len(s) > fs.window {
		s = s[len(s)-fs.window:]
	}
	fs.data[nodeID] = s
}

func (fs *FeatureStore) Get(nodeID int) []float64 {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	s := fs.data[nodeID]
	out := make([]float64, len(s))
	copy(out, s)
	return out
}

func Predict(series []float64) float64 {
	n := len(series)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return series[0]
	}
	var sx, sy, sxx, sxy float64
	for i := 0; i < n; i++ {
		x, y := float64(i), series[i]
		sx += x
		sy += y
		sxx += x * x
		sxy += x * y
	}
	den := float64(n)*sxx - sx*sx
	if math.Abs(den) < 1e-9 {
		return sy / float64(n)
	}
	a := (float64(n)*sxy - sx*sy) / den
	b := (sy - a*sx) / float64(n)
	p := a*float64(n) + b
	if p < 0 {
		p = 0
	}
	return p
}

func monitor(node *Node, out chan<- [2]int, stop <-chan struct{}) {
	t := time.NewTicker(300 * time.Millisecond)
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			node.mu.Lock()
			node.UsedMB += rand.Intn(81) - 40
			if node.UsedMB < 0 {
				node.UsedMB = 0
			}
			if node.UsedMB > node.TotalMB {
				node.UsedMB = node.TotalMB
			}
			out <- [2]int{node.ID, node.UsedMB}
			node.mu.Unlock()
		}
	}
}

func controller(nodes []*Node, metrics <-chan [2]int, fs *FeatureStore, stop <-chan struct{}) {
	t := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-stop:
			return
		case m := <-metrics:
			fs.Add(m[0], m[1])
		case <-t.C:
			for _, n := range nodes {
				s := fs.Get(n.ID)
				p := Predict(s)
				r := int(math.Ceil(p * 1.2))
				n.mu.Lock()
				if r < n.UsedMB {
					r = n.UsedMB
				}
				if r > 8192 {
					r = 8192
				}
				if r < 512 {
					r = 512
				}
				n.TotalMB = r
				n.mu.Unlock()
			}
		}
	}
}

func printStatus(nodes []*Node) {
	fmt.Println("----- Cluster Status -----")
	for _, n := range nodes {
		n.mu.Lock()
		u := float64(n.UsedMB) / float64(n.TotalMB) * 100
		fmt.Printf("Node %02d | Total:%5dMB | Used:%5dMB | Util:%5.1f%%\n", n.ID, n.TotalMB, n.UsedMB, u)
		n.mu.Unlock()
	}
}

func createNodes(c int) []*Node {
	nodes := make([]*Node, c)
	for i := 0; i < c; i++ {
		t := 1024 + rand.Intn(1024)
		u := 256 + rand.Intn(t/2)
		nodes[i] = &Node{ID: i + 1, TotalMB: t, UsedMB: u}
	}
	return nodes
}

func main() {
	rand.Seed(time.Now().UnixNano())
	nodes := createNodes(5)
	metrics := make(chan [2]int, 1024)
	stop := make(chan struct{})
	fs := NewFeatureStore(10)
	for _, n := range nodes {
		go monitor(n, metrics, stop)
	}
	go controller(nodes, metrics, fs, stop)
	ticker := time.NewTicker(3 * time.Second)
	done := time.After(30 * time.Second)
	for {
		select {
		case <-done:
			close(stop)
			time.Sleep(500 * time.Millisecond)
			printStatus(nodes)
			return
		case <-ticker.C:
			printStatus(nodes)
		}
	}
}

