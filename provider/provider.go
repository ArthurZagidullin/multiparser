package provider

// Instance среда выполнение ssh команд
type Instance interface {
	Execute()
}

// Provider задача подготовить instance
type Provider interface {
	GetInstance() chan <-Instance
}
