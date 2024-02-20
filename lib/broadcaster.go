package lib

import "context"

type Broadcaster[K interface{}] struct {
	source    chan K
	listeners []chan K
}

func (b *Broadcaster[K]) Serve(source chan K, ctx context.Context) {
	b.listeners = make([]chan K, 0)
	b.source = source
	defer func() {
		for _, v := range b.listeners {
			close(v)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-source:
			{
				for _, v := range b.listeners {
					v <- msg
				}
			}
		}
	}
}

func (b *Broadcaster[K]) Subscribe() chan K {
	newChan := make(chan K)

	b.listeners = append(b.listeners, newChan)
	return newChan
}

func (b *Broadcaster[K]) Unsubscribe(listener chan K) {
	for k, v := range b.listeners {
		if v == listener {
			left := b.listeners[:k]
			right := b.listeners[k+1:]
			b.listeners = append(left, right...)
		}
	}
}
