package model

const (
	StatusNormal = 1
	Personal     = 1
)

const (
	AESKey = "asdfghjklzxcvbnmqwertyui"
)

const (
	NoArchive = iota
	Archive
)

const (
	NoDeleted = iota
	Deleted
)

const (
	Open = iota
	Private
	Custom
)

const (
	Default = "default"
	Simple  = "simple"
)

const (
	NoCollected = iota
	Collected
)

const (
	NoOwner = iota
	Owner
)

const (
	NoExecutor = iota
	Executor
)

const (
	NoCanRead = iota
	CanRead
)

const (
	UnDone = iota
	Done
)

const (
	MyExecute = iota + 1
	MyParticipate
	MyCreate
)

const (
	NoComment = iota
	Comment
)

const (
	Active = iota + 1
	System
	Baned
	ActiveAndCurDept
)
