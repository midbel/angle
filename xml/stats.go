package xml

import (
	"io"
)

type Stats struct {
	MaxDepth   int
	Depth      map[int]int
	Namespaces map[Namespace]struct{}

	Elements       map[string]int
	QualifiedEls   map[string]int
	Attributes     map[string]int
	QualifiedAttrs map[string]int
	Types          map[NodeType]int
	Paths          map[string]int
}

func Infos(r io.Reader) (*Stats, error) {
	var (
		rs = NewReader(r)
		sh = createStatHandler()
	)
	if err := rs.Read(sh); err != nil {
		return nil, err
	}
	return sh.stats, nil
}

type statsHandler struct {
	stats *Stats
	path  []string
	depth int
}

func createStatHandler() *statsHandler {
	stats := &Stats{
		Namespaces:     make(map[Namespace]struct{}),
		Elements:       make(map[string]int),
		QualifiedEls:   make(map[string]int),
		Attributes:     make(map[string]int),
		QualifiedAttrs: make(map[string]int),
		Types:          make(map[NodeType]int),
		Depth:          make(map[int]int),
		Paths:          make(map[string]int),
	}
	return &statsHandler{
		stats: stats,
	}
}

func (h *statsHandler) OnStartElement(e Element, _ bool) error {
	h.stats.Elements[e.LocalName()]++
	h.stats.QualifiedEls[e.QualifiedName()]++
	h.stats.Types[ElementNode]++

	h.path = append(h.path, e.LocalName())

	for _, a := range e.Attributes {
		h.stats.Types[AttrNode]++
		if a.Prefix == "xmlns" {
			ns := Namespace{
				Prefix: "",
				URI:    a.Value,
			}
			if _, ok := h.stats.Namespaces[ns]; !ok {
				h.stats.Namespaces[ns] = struct{}{}
			}
		} else {
			h.stats.Attributes[a.LocalName()]++
			h.stats.QualifiedAttrs[a.QualifiedName()]++
		}
	}
	h.depth++
	h.stats.Depth[h.depth]++
	return nil
}

func (h *statsHandler) OnCloseElement(_ Name) error {
	h.stats.MaxDepth = max(h.depth, h.stats.MaxDepth)
	h.depth--
	if n := len(h.path); n > 0 {
		h.path = h.path[:n-1]
	}
	return nil
}

func (h *statsHandler) OnText(Text) error {
	h.stats.Types[TextNode]++
	return nil
}

func (h *statsHandler) OnComment(Comment) error {
	return nil
}

func (h *statsHandler) OnPI(PI) error {
	return nil
}
