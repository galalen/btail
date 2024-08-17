package main

type LogBufferQueue struct {
	data []Line
	size int
}

func NewLogBufferQueue(size int) *LogBufferQueue {
	return &LogBufferQueue{
		data: make([]Line, 0, size),
		size: size,
	}
}

func (q *LogBufferQueue) Push(item Line) {
	if len(q.data) == q.size {
		q.data = q.data[1:]
	}
	q.data = append(q.data, item)
}

func (q *LogBufferQueue) GetAll() []Line {
	return q.data
}
