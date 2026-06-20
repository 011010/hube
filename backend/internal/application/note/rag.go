package note

import (
	"context"
	"encoding/json"
	"log"
	"sort"

	"github.com/husari/hube/internal/domain/note"
	"github.com/husari/hube/internal/infrastructure/external"
)

type RAGService struct {
	repo       note.EmbeddingRepository
	embeddings *external.EmbeddingsClient
}

func NewRAGService(repo note.EmbeddingRepository, embeddings *external.EmbeddingsClient) *RAGService {
	return &RAGService{repo: repo, embeddings: embeddings}
}

// IndexNote generates and stores an embedding for the given note.
func (s *RAGService) IndexNote(ctx context.Context, n *note.Note) {
	text := n.Title + "\n" + n.Content
	vec, err := s.embeddings.Embed(ctx, text)
	if err != nil {
		log.Printf("rag: embed note %s: %v", n.ID, err)
		return
	}
	data, err := json.Marshal(vec)
	if err != nil {
		log.Printf("rag: marshal embedding %s: %v", n.ID, err)
		return
	}
	if err := s.repo.StoreEmbedding(ctx, n.ID, data); err != nil {
		log.Printf("rag: store embedding %s: %v", n.ID, err)
	}
}

type SemanticResult struct {
	Note  note.Note `json:"note"`
	Score float32   `json:"score"`
}

// SemanticSearch returns top-k notes most similar to the query.
func (s *RAGService) SemanticSearch(ctx context.Context, query string, topK int) ([]SemanticResult, error) {
	queryVec, err := s.embeddings.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	records, err := s.repo.FindAllEmbeddings(ctx)
	if err != nil {
		return nil, err
	}

	type scored struct {
		id    string
		score float32
	}
	scores := make([]scored, 0, len(records))
	for _, rec := range records {
		var vec []float32
		if err := json.Unmarshal(rec.Embedding, &vec); err != nil {
			continue
		}
		scores = append(scores, scored{id: rec.ID, score: external.CosineSimilarity(queryVec, vec)})
	}

	sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })
	if topK > 0 && len(scores) > topK {
		scores = scores[:topK]
	}

	results := make([]SemanticResult, 0, len(scores))
	for _, sc := range scores {
		if sc.score < 0.3 {
			continue
		}
		n, err := s.repo.FindByID(ctx, sc.id)
		if err != nil || n == nil {
			continue
		}
		results = append(results, SemanticResult{Note: *n, Score: sc.score})
	}
	return results, nil
}
