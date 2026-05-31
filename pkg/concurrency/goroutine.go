package concurrency

// Định nghĩa orDone (Gộp kênh) để  lắng nghe tín hiệu kết thúc từ nhiều goroutine khác nhau
func orDone(channels ...<-chan struct{}) <-chan struct{} {
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}

	orChan := make(chan struct{})
	go func() {
		defer close(orChan)
		switch len(channels) {
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			select {
			case <-channels[0]:
			case <-channels[1]:
			case <-orDone(append(channels[2:], orChan)...):
			}
		}
	}()

	return orChan
}

// Định nghĩa Bridge để duỗi thẳng các chan trong chan thành 1 chan data

// Định nghĩa Stateful Ward (Ward nhưng store được trạng thái Ward trước kia đã làm gì)

// Định nghĩa Steward (Giám sát Heartbeat của Ward để ra quyết định clear and new)
