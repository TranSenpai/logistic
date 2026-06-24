package concurrencyUtils

import (
	"context"
)

// Gom N tín hiệu hủy thành tín hiệu hủy tổng
func OrChannel(cancelSignal ...<-chan struct{}) <-chan struct{} {
	switch len(cancelSignal) {
	case 0:
		return nil
	case 1:
		return cancelSignal[0]
	}

	finalCancelSignal := make(chan struct{})
	go func() {
		defer close(finalCancelSignal)
		switch len(cancelSignal) {
		case 2:
			select {
			case <-cancelSignal[0]:
			case <-cancelSignal[1]:
			}
		default:
			select {
			case <-cancelSignal[0]:
			case <-cancelSignal[1]:
			case <-cancelSignal[2]:
			case <-OrChannel(append(cancelSignal[3:], finalCancelSignal)...):
			}
		}
	}()

	return finalCancelSignal
}

// Đọc một channels hỗ trợ khả năng ngắt ngang
func OrDone[T any](ctx context.Context, channel <-chan T) <-chan T {
	valStream := make(chan T)
	go func() {
		defer close(valStream)
		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-channel:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case valStream <- val:
				}
			}
		}
	}()

	return valStream
}

// Định nghĩa Bridge để duỗi thẳng các chan trong chan thành 1 chan data
func Bridge[T any](ctx context.Context, channels <-chan <-chan T) <-chan T {
	valStream := make(chan T)
	go func() {
		defer close(valStream)
		for {
			var stream <-chan T
			select {
			case <-ctx.Done():
				return
			case subChannels, ok := <-channels:
				if !ok {
					return
				}
				stream = subChannels
			}

			for val := range OrDone(ctx, stream) {
				select {
				case <-ctx.Done():
					return
				case valStream <- val:
				}
			}
		}
	}()

	return valStream
}

type ChanStruct[U any] struct {
	value U
	err   error
}

func (p ChanStruct[U]) GetError() error {
	return p.err
}

func (p ChanStruct[U]) GetResult() (U, error) {
	if p.err != nil {
		var nilReuslt U
		return nilReuslt, p.err
	}

	return p.value, nil
}

func Pipeline[T, U any](ctx context.Context, inStream <-chan T, fn func(ctx context.Context, input T) (U, error)) <-chan ChanStruct[U] {
	outStream := make(chan ChanStruct[U])
	go func() {
		defer close(outStream)

		for val := range inStream {
			select {
			case <-ctx.Done():
				return
			default:
			}

			result, err := fn(ctx, val)
			select {
			case <-ctx.Done():
				return
			case outStream <- ChanStruct[U]{value: result, err: err}:
			}
		}
	}()

	return outStream
}

func primeFinder(ctx context.Context, intStream <-chan int) <-chan int {
	primeStream := make(chan int)
	go func() {
		defer close(primeStream)
		for integer := range intStream {
			// Thuật toán kiểm tra số nguyên tố ngây ngô và cực chậm
			prime := true
			for i := 2; i < integer; i++ {
				if integer%i == 0 {
					prime = false
					break
				}
			}

			// Nếu là số nguyên tố, đẩy vào ống đầu ra
			if prime {
				select {
				case <-ctx.Done():
					return
				case primeStream <- integer:
				}
			}
		}
	}()
	return primeStream
}

func toInt(ctx context.Context, valueStream <-chan any) <-chan int {
	intStream := make(chan int)

	go func() {
		defer close(intStream)
		for v := range valueStream {
			select {
			case <-ctx.Done():
				return
			case intStream <- v.(int):
			}
		}
	}()

	return intStream
}

func repeatFn(ctx context.Context, fn func() any) <-chan any {
	valueStream := make(chan any)

	go func() {
		defer close(valueStream)
		for {
			select {
			case <-ctx.Done():
				return
			case valueStream <- fn():
			}
		}
	}()

	return valueStream
}

func take(ctx context.Context, valueStream <-chan string, num int) <-chan string {
	takeStream := make(chan string)

	go func() {
		defer close(takeStream)

		for i := num; i > 0 || i == -1; {
			if i != -1 {
				i--
			}
			select {
			case <-ctx.Done():
				return
			case takeStream <- <-valueStream:
			}
		}
	}()

	return takeStream
}

// func FanIn()

// func FanOutFanIn(ctx context.Context) {
// 	randFn := func() any {
// 		return rand.IntN(50000000)
// 	}

// 	start := time.Now()
// 	numFinders := runtime.NumCPU()

// 	fmt.Printf("Spinning up %d prime finders.\n", numFinders)
// 	finders := make([]<-chan interface{}, numFinders)
// 	fmt.Println("Primes:")

// 	for i := 0; i < numFinders; i++ {
// 		finders[i] = primeFinder(done, randIntStream)
// 	}

// 	for prime := range take(done, fanIn(done, finders...), 10) {
// 		fmt.Printf("\t%d\n", prime)
// 	}
// 	fmt.Printf("Search took: %v", time.Since(start))
// }

// // func FanOut[T, U any](ctx context.Context, fn func(ctx context.Context, input T) (U error)) (<-chan ChanStruct[U], error) {
// 	outStream := make(chan U)

// loop:
// 	for {
// 		go func() {
// 			select {
// 			case <-ctx.Done():
// 				break loop
// 			default:

// 			}
// 		}()
// 	}

// 	return outStream, nil
// }

// Định nghĩa Stateful Ward (Ward nhưng store được trạng thái Ward trước kia đã làm gì)

// Định nghĩa Steward (Giám sát Heartbeat của Ward để ra quyết định clear and new)
