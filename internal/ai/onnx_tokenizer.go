package ai

import (
    "strings"
)

type SimpleTokenizer struct {
    vocab map[string]int
}

func NewSimpleTokenizer() *SimpleTokenizer {
    // Simple vocabulary - in production, load from file
    vocab := map[string]int{
        "[PAD]": 0,
        "[UNK]": 1,
        "[CLS]": 2,
        "[SEP]": 3,
        // Add more tokens as needed
    }
    
    return &SimpleTokenizer{vocab: vocab}
}

func (t *SimpleTokenizer) Encode(text string) []int {
    tokens := strings.Fields(strings.ToLower(text))
    ids := []int{t.vocab["[CLS]"]} // Start token
    
    for _, token := range tokens {
        if id, ok := t.vocab[token]; ok {
            ids = append(ids, id)
        } else {
            ids = append(ids, t.vocab["[UNK]"])
        }
    }
    
    ids = append(ids, t.vocab["[SEP]"]) // End token
    
    return ids
}

func (t *SimpleTokenizer) PadSequence(ids []int, maxLength int) []int {
    if len(ids) > maxLength {
        return ids[:maxLength]
    }
    
    padded := make([]int, maxLength)
    copy(padded, ids)
    
    return padded
}