package service

import (
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
)

func TestGraphExtractMarkerRoundTrip(t *testing.T) {
	content := "测试内容 chunk body v1"
	cfg := types.ExtractConfig{Tags: []string{"依赖于"}, Text: "prompt", CustomInstructions: "custom"}
	modelID := "model-a"

	fp := graphExtractConfigFingerprint(modelID, cfg)
	hash := graphExtractContentHash(content)

	// simulate metadata as persisted: {"graph_extract": {...}}
	raw := types.JSON(`{"graph_extract": {"fingerprint": "` + fp + `", "content_hash": "` + hash + `"}}`)
	m := readGraphExtractMarker(raw)
	if m == nil {
		t.Fatal("marker not read back")
	}
	if prevFp, _ := m["fingerprint"].(string); prevFp != fp {
		t.Fatalf("fingerprint mismatch: %v", prevFp)
	}
	if prevHash, _ := m["content_hash"].(string); prevHash != hash {
		t.Fatalf("content hash mismatch: %v", prevHash)
	}

	// unchanged content + config => skip conditions all match
	if graphExtractContentHash(content) != hash {
		t.Fatal("same content produced different hash")
	}
	// changed content => hash differs => no skip
	if graphExtractContentHash(content+" changed") == hash {
		t.Fatal("changed content produced same hash")
	}
	// changed config => fingerprint differs => no skip
	cfg2 := cfg
	cfg2.Tags = append(cfg2.Tags, "包含")
	if graphExtractConfigFingerprint(modelID, cfg2) == fp {
		t.Fatal("changed config produced same fingerprint")
	}
	// changed model => fingerprint differs
	if graphExtractConfigFingerprint("model-b", cfg) == fp {
		t.Fatal("changed model produced same fingerprint")
	}
	// empty metadata => nil marker (no skip)
	if readGraphExtractMarker(types.JSON(`{}`)) != nil {
		t.Fatal("empty metadata should yield nil marker")
	}
}
