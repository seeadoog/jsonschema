package expr

import (
	"context"
	"sync"
	"testing"
)

func TestCache(t *testing.T) {
	ic := 0
	c := NewInstanceCache[string, any, any](func(ctx context.Context, k string, c any) (v any, err error) {
		ic++
		return ic, nil
	})

	var v any
	var err error
	wg := &sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err = c.Get(context.Background(), "1", nil)
			if err != nil {
				t.Fatal(err)
			}
		}()
	}
	wg.Wait()

	assertEqual2(t, v, 1)
	assertEqual2(t, err, nil)
}
