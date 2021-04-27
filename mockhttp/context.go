package mockhttp

import (
	"context"
	"fmt"
	"sort"
)

type conversationScores struct {
	values map[string]int
}
type scoreMap map[string]int
type sessionIdKey struct{}
type scoreKey struct{}

func contextWithWithSessionId(id int) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, scoreKey{}, &conversationScores{values: make(scoreMap)})
	return context.WithValue(ctx, sessionIdKey{}, id)
}

func getSessionId(ctx context.Context) (int, error) {
	return getContextValueAsInt(ctx, sessionIdKey{})
}

func getConversationScores(ctx context.Context) (*conversationScores, error) {
	if v := ctx.Value(scoreKey{}); v != nil {
		if t, ok := v.(*conversationScores); ok {
			return t, nil
		} else {
			return nil, fmt.Errorf("invalid value type for %T", scoreKey{})
		}
	} else {
		return nil, fmt.Errorf("no value found for %T", scoreKey{})
	}
}

func getContextValueAsInt(ctx context.Context, key interface{}) (int, error) {
	if v := ctx.Value(key); v != nil {
		if t, ok := v.(int); ok {
			return t, nil
		} else {
			return 0, fmt.Errorf("invalid value type for %T", key)
		}
	} else {
		return 0, fmt.Errorf("no value found for %T", key)
	}
}

func (s *conversationScores) inc(name string) {
	if v, found := s.values[name]; found {
		s.values[name] = v + 1
	} else {
		s.values[name] = 1
	}
}

func (s *conversationScores) tieBreak(candidates []Conversation) (Conversation, error) {
	type kv struct {
		k string
		v int
	}
	var ss []kv
	for k, v := range s.values {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].v > ss[j].v
	})

	for _, s := range ss {
		for _, c := range candidates {
			if s.k == c.Name {
				return c, nil
			}
		}
	}
	return Conversation{}, fmt.Errorf("that is wierd, no candidates found in score")
}
