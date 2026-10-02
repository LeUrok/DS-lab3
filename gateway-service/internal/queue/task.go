package queue

type TaskType string

const (
	TaskCancelPayment TaskType = "cancel_payment"
	TaskUnreserveCar  TaskType = "unreserve_car"
)

type Task struct {
	Type       TaskType
	Payload    map[string]string
	Retries    int
	MaxRetries int
}