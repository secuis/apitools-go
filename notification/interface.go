package notification

type Sender interface {
	Alert(msg string)
	Info(msg string)
}
