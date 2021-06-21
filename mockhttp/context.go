package mockhttp

import (
	"context"
	"fmt"
	"github.com/thorsager/mockdev/logging"
	"sort"
)

type conversationScores struct {
	values map[string]int
}
type scoreMap map[string]int
type sessionIdKey struct{}
type scoreKey struct{}
type logKey struct{}

func contextWithWithSessionId(id int) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, scoreKey{}, &conversationScores{values: make(scoreMap)})
	return context.WithValue(ctx, sessionIdKey{}, id)
}
func setContextLogger(ctx context.Context, logger logging.Logger) context.Context {
	return context.WithValue(ctx, logKey{}, logger)
}
func getContextLogger(ctx context.Context) (logging.Logger, error) {
	if v := ctx.Value(logKey{}); v != nil {
		if t, ok := v.(logging.Logger); ok {
			return t, nil
		} else {
			return nil, fmt.Errorf("invalid value type for %T", logKey{})
		}
	} else {
		return nil, fmt.Errorf("no value found for %T", logKey{})
	}
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

func (s *conversationScores) bump(name string, cnt int) {
	if v, found := s.values[name]; found {
		s.values[name] = v + cnt
	} else {
		s.values[name] = cnt
	}
}

func (s *conversationScores) inc(name string) {
	s.bump(name, 1)
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
